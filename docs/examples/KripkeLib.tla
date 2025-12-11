---- MODULE KripkeLib ----
\* Standard library for kripke-ctl TLA+ specifications
\* Defines operators for process calculus notation and probabilistic choice

EXTENDS Naturals, Sequences, FiniteSets

\* ============================================================================
\* CHANNEL OPERATIONS (Process Calculus Style)
\* ============================================================================

\* Send: channel ! message
\* Returns the channel with message appended
snd(channel, message) == Append(channel, message)

\* Receive: channel ? variable
\* Returns the first message and the updated channel
\* Usage: LET result == rcv(channel) 
\*        IN msg' = result.msg /\ channel' = result.channel
rcv(channel) == 
    [msg |-> Head(channel), channel |-> Tail(channel)]

\* Check if channel can send (not full)
can_send(channel, capacity) == Len(channel) < capacity

\* Check if channel can receive (not empty)
can_recv(channel) == Len(channel) > 0

\* ============================================================================
\* PROBABILISTIC CHOICE
\* ============================================================================

\* choice(lower <= R < upper, guard, action)
\* Represents a probabilistic branch with INTEGER PERCENTAGES (0-100)
\*
\* Semantics:
\*   - R is a random variable ~ UniformInt[0,100) rolled on each scheduler attempt
\*   - If lower <= R < upper, this branch is enabled
\*   - guard is an additional condition that must hold
\*   - action is the state update
\*
\* Example:
\*   choice(0, 70, TRUE, action1)    \* 70% probability
\*   choice(70, 100, TRUE, action2)  \* 30% probability
\*
\* In TLA+ model checking:
\*   - TLC treats this as non-deterministic choice (explores all branches)
\*   - The probability annotation is preserved for documentation
\*
\* In simulation:
\*   - Roll R ~ UniformInt[0,100)
\*   - Evaluate each choice(lower, upper, guard, action)
\*   - Take the first one where lower <= R < upper AND guard is true
\*
\* Note: Multiple choice operators in a disjunction should have:
\*   - Non-overlapping ranges (sum to 100)
\*   - Same R variable (e.g., all use R1)
\*
choice(lower, upper, guard, action) ==
    \* For TLA+: just evaluate guard /\ action
    \* The probability range [lower, upper) is documentation
    guard /\ action

\* Helper: Create a probabilistic choice between two actions
\* prob_choice(p, action1, action2) means:
\*   - action1 with probability p%
\*   - action2 with probability (100-p)%
prob_choice(p, action1, action2) ==
    \/ choice(0, p, TRUE, action1)
    \/ choice(p, 100, TRUE, action2)

\* ============================================================================
\* CHANNEL TYPES
\* ============================================================================

\* A channel is a sequence of messages
Channel == Seq(Nat)

\* A bounded channel has a maximum capacity
BoundedChannel(capacity) == {c \in Channel : Len(c) <= capacity}

====
