---
name: tdd
description: "acta: Use for every code change - feature, bug fix, refactor, config - before writing implementation code. Red, green, refactor - a failing test first, watched failing, then the minimum code, watched passing."
---

# Test-Driven Development (TDD)

Write the test first. Watch it fail. Write minimal code to pass.

**Core principle:** If you didn't watch the test fail, you don't know if it tests the right thing. Violating the letter of the rules is violating the spirit of the rules.

## When to Use

Always: new features, bug fixes, refactoring, behavior changes.
Exceptions (ask your human partner): throwaway prototypes, generated code. "Skip TDD just this once" is rationalization.

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over. Don't keep it as "reference", don't "adapt" it, don't look at it. Implement fresh from tests.

## Red-Green-Refactor

### RED - Write Failing Test

Write one minimal test: one behavior, a clear name, real code (no mocks unless unavoidable).

<Good>
```typescript
test('retries failed operations 3 times', async () => {
  let attempts = 0;
  const op = () => {
    if (++attempts < 3) throw new Error('fail');
    return 'success';
  };
  expect(await retryOperation(op)).toBe('success');
  expect(attempts).toBe(3);
});
```
Clear name, real behavior
</Good>

<Bad>
```typescript
test('retry works', async () => {
  const mock = jest.fn().mockRejectedValueOnce(new Error());
  await retryOperation(mock);
  expect(mock).toHaveBeenCalledTimes(2);
});
```
Vague name, tests the mock
</Bad>

### Verify RED - Watch It Fail

**MANDATORY. Never skip.** Run the test. Confirm it fails (not errors), the message is expected, and it fails because the feature is missing (not typos).

Test passes? You're testing existing behavior; fix it. Test errors? Fix the error, re-run until it fails correctly.

### GREEN - Minimal Code

Write the simplest code that passes the test. Don't add features, refactor other code, or "improve" beyond the test.

<Good>
```typescript
async function retryOperation<T>(fn: () => Promise<T>): Promise<T> {
  for (let i = 0; ; i++) {
    try { return await fn(); } catch (e) { if (i === 2) throw e; }
  }
}
```
Just enough
</Good>

<Bad>
```typescript
async function retryOperation<T>(fn: () => Promise<T>,
  options?: { maxRetries?: number; backoff?: 'linear' | 'exponential' }): Promise<T> {
  // YAGNI
}
```
Over-engineered
</Bad>

### Verify GREEN - Watch It Pass

**MANDATORY.** Confirm the test passes, other tests still pass, output is pristine (no errors, warnings). Test fails? Fix the code, not the test. Other tests fail? Fix now. The full suite runs only in `acta:land`.

### REFACTOR - Clean Up

After green only: remove duplication, improve names, extract helpers. Keep tests green, add no behavior. Then the next failing test.

## Good Tests

- **Minimal:** one thing. "and" in the name? Split it.
- **Clear:** the name describes behavior, never `test1`.
- **Shows intent:** it demonstrates the desired API.

When writing or changing any test, read the rules that keep tests honest in [writing-good-tests.md](writing-good-tests.md):
- Name the production change that would make the test fail — before writing it
- Assert on real behavior, never on mock behavior
- Keep test-only code in test utilities, out of production classes
- Understand a dependency's side effects before mocking it

## Cover the plan, not the universe

Write the plan's cases and no others: every branch, error path, validation rule and guard the plan names gets one test, and you test it once, at the layer that owns it. A layer above tests wiring only: one success path and one error passed up from below. Never prove a lower layer's cases again.

Test through the seam the plan names, or with none named the public interface callers use, never private code. Inputs that share one code path go in one table-driven test, one row each. One test, then its code, then the next. Never write all the tests before any code.

Adversarial inputs go only where the input crosses a trust boundary: user input, files, network, auth, money, migrations, deletes. There the plan names which adversarial inputs to try, and you write the ones it names. Kinds: empty, null, 0, negative, huge, unicode, off-by-one both sides, wildcard and escape characters, wrong-domain value.

A bug fix gets a regression test that reproduces the exact bug and fails before the fix. A path the plan does not cover is a note for a later plan, not a test now.

## Quality bar

- A test that stays green with the implementation reverted is not a test. Check that it goes red.
- Assert effects, not mocks. Mock only the collaborator the change touches. Cover the flag-off path of every flag the change adds.
- Test tooling (fixtures, probes, cleanup helpers) only serves the deliverable: minimum that works, no review rounds of its own.
- "Obvious", "one-liner" and "just config" are not exemptions.

## Rationalizations - STOP and Start Over

Each means: delete the code, start over with a test first.

- Code before the test, or "tests after reach the same goal". Tests written after prove nothing.
- Skipping the red run, a test that passes at once, or a failure you cannot explain.
- "Too simple", "just this once", "I already tested it by hand", "I will add tests later". Manual checks cannot be re-run.
- "Keep it as reference" or "adapt it". That is testing after.
- "Deleting X hours of work is wasteful". Sunk cost; code you cannot trust is the waste.
- "Need to explore first". Fine. Throw the exploration away, then write the test.
- "Hard to test" means the design is unclear. Simplify the interface.
- "Existing code has no tests". Test the behavior you change; the rest is a note for a later plan.
- "TDD slows me down" or "this is different". Skipping it means debugging later, which is slower.

## When Stuck

| Problem | Solution |
|---------|----------|
| Don't know how to test | Write the wished-for API and the assertion first. Ask your human partner. |
| Test too complicated | Design too complicated. Simplify the interface. |
| Must mock everything | Code too coupled. Use dependency injection. |
| Test setup huge | Extract helpers. Still complex? Simplify design. |

## Debugging Integration

Bug found? Write a failing test that reproduces it and follow the cycle. Never fix a bug without a test, unless your human partner allows it.
