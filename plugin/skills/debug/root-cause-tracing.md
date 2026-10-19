# Root Cause Tracing

A bug often shows up deep in a call stack: `git init` in the wrong folder, a
file written to the wrong place, a database opened with the wrong path. The
place it shows up is rarely where it starts. Trace the bad value back to its
source and fix it there.

## When to use

- The error happens deep in execution, not at the entry point.
- The stack trace shows a long call chain.
- It is unclear where the bad value came from.
- You need to know which test or caller triggers it.

If you cannot trace further back, say where the trail ends and what you
could not see. Do not guess the rest of the chain.

## The tracing steps

1. **Observe the symptom.**
   `Error: git init failed in ~/project/packages/core`
2. **Find the line that does it.**
   `await execFileAsync('git', ['init'], { cwd: projectDir });`
3. **Ask what called it, with what value.**
   ```
   WorktreeManager.createSessionWorktree(projectDir, sessionId)
     called by Session.initializeWorkspace()
     called by Session.create()
     called by a test at Project.create()
   ```
   `projectDir` was `''`. An empty `cwd` falls back to `process.cwd()`, which
   is the source folder.
4. **Keep going up until the value is made.**
   ```typescript
   const context = setupCoreTest(); // gives { tempDir: '' } until setup runs
   Project.create('name', context.tempDir); // read before beforeEach filled it
   ```
5. **Fix at the source.** Here: make `tempDir` throw if read before setup.
   Then decide if a check belongs anywhere else: see `defense-in-depth.md`.

## When you cannot trace by reading

Log right before the risky call, in a temp copy outside the repo:

```typescript
async function gitInit(directory: string) {
  console.error('DEBUG git init:', {
    directory,
    cwd: process.cwd(),
    stack: new Error().stack,
  });
  await execFileAsync('git', ['init'], { cwd: directory });
}
```

Run it and keep only those lines:

```bash
npm test 2>&1 | grep 'DEBUG git init'
```

Then read the stacks: which test file, which line, which value repeats.

- In tests, use `console.error()`. A logger may be switched off there.
- Log before the risky call, not after it fails.
- Log paths, the working folder and the values passed in. Never log secret
  values: print whether a secret is set, not what it is.
- Remove the logging when you are done.

## Which test leaves state behind

If a test leaves a file or folder behind and you do not know which one, the
script `find-polluter.sh` in this folder runs the tests one by one and stops
at the first that creates it:

```bash
./find-polluter.sh '.git' 'src/**/*.test.ts'
```

For a test that fails only after another test runs, see the order-dependent
recipe in `SKILL.md`.
