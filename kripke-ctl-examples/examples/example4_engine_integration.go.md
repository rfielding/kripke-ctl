===============================================================================
 Producer-Consumer using kripke engine + CTL verification
===============================================================================

PART 1: Running the Actor Engine
------------------------------------------------------------------------------
Running 10 steps...
  Step 1: Time=1, Buffer=1, Events=0
  Step 2: Time=2, Buffer=0, Events=1
  Step 3: Time=3, Buffer=1, Events=1
  Step 4: Time=4, Buffer=2, Events=1
  Step 5: Time=5, Buffer=1, Events=2
  Step 6: Time=6, Buffer=0, Events=3
  Step 7: Time=7, Buffer=1, Events=3
  Step 8: Time=8, Buffer=2, Events=3
  Step 9: Time=9, Buffer=1, Events=4
  Step 10: Time=10, Buffer=0, Events=5

Final state: Buffer=0, Consumer received=5 messages

PART 2: Building State Space Graph
------------------------------------------------------------------------------
States (3 total):
  buffer_0: P=true C=false
  buffer_1: P=true C=true
  buffer_2: P=false C=true

PART 3: CTL Model Checking
------------------------------------------------------------------------------
✓ PASS Safety: Buffer never overflows
✓ PASS Liveness-P: Producer can always eventually send
✓ PASS Liveness-C: Consumer can always eventually receive
✓ PASS No-Deadlock: System never deadlocks
✓ PASS Reachability-Full: Buffer can become full
✓ PASS Reachability-Empty: Buffer can become empty

PART 4: Mermaid State Diagram
```mermaid
stateDiagram-v2
    [*] --> buffer_0

    buffer_0 --> buffer_1: produce
    buffer_1 --> buffer_2: produce
    buffer_1 --> buffer_0: consume
    buffer_2 --> buffer_1: consume

    buffer_0: buffer_0 (P:ready C:blocked)
    buffer_1: buffer_1 (P:ready C:ready)
    buffer_2: buffer_2 (P:blocked C:ready)


PART 5: Event Log (Queue Delay Analysis)
------------------------------------------------------------------------------
  Msg 1: Delay=1 ticks, Time=1
  Msg 2: Delay=2 ticks, Time=4
  Msg 3: Delay=2 ticks, Time=5
  Msg 4: Delay=2 ticks, Time=8
  Msg 5: Delay=2 ticks, Time=9

Average queue delay: 1.80 ticks
```
