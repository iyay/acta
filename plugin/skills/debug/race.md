# Forcing a race

Repeat runs only make a lost update likely. This probe makes it happen on
every run: both workers read, wait at a barrier until the other has read too,
and only then write. Run it in a temp copy, never in the repo.

Python, for a read-then-write method. Change the names to the real ones:

```python
import threading
from app.store import Store          # the class with the racy method

real_read = Store.read
gate = threading.Barrier(2)

def held_read(self):
    value = real_read(self)
    try:
        gate.wait(timeout=5)          # both workers have read the same value
    except threading.BrokenBarrierError:
        pass                          # a lock made the other worker wait
    return value

Store.read = held_read
store = Store("/tmp/probe/state.json")
workers = [threading.Thread(target=store.increment) for _ in range(2)]
for w in workers: w.start()
for w in workers: w.join()
print("expected 2, got", real_read(store))   # 1 means a lost update
```

Other languages: the same shape with their own barrier or latch, such as
`sync.WaitGroup` plus a channel in Go or `CountDownLatch` in Java.

Report the printed line as the evidence. A lock around read and write in the
same copy, printing 2, is the control.

The regression test uses the same barrier on the read, runs the two workers,
and asserts the final value equals the number of updates. A fixed version must
pass it: once the method holds a lock, the second worker's read waits, so the
barrier times out and the probe goes on.
