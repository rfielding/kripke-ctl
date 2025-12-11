---- MODULE ProducerConsumer ----
EXTENDS Naturals, Sequences, FiniteSets, KripkeLib

CONSTANTS
    MaxSteps,      \* Maximum execution steps
    Cap_consumer_inbox,  \* Channel capacity
    MaxMessages    \* Maximum messages to send

VARIABLES
    producer_count,      \* producer message count
    consumer_count,      \* consumer message count
    chan_consumer_inbox,         \* Channel buffer (sequence)
    step_count     \* Total execution steps

vars == <<producer_count, consumer_count, chan_consumer_inbox, step_count>>

TypeOK ==
    /\ producer_count \in Nat
    /\ consumer_count \in Nat
    /\ chan_consumer_inbox \in Seq(Nat)
    /\ step_count \in Nat

Init ==
    /\ producer_count = 0
    /\ consumer_count = 0
    /\ chan_consumer_inbox = <<>>
    /\ step_count = 0

\* Actor: producer
\* Process calculus: producer.out ! msg
producer_Send ==
    /\ producer_count < MaxMessages      \* Guard: not done sending
    /\ can_send(chan_consumer_inbox, Cap_consumer_inbox)  \* Guard: channel not full
    /\ chan_consumer_inbox' = snd(chan_consumer_inbox, producer_count)  \* producer.out ! msg
    /\ producer_count' = producer_count + 1           \* Update: count' = count + 1
    /\ UNCHANGED <<consumer_count, step_count>>

\* Actor: consumer
\* Process calculus: consumer.inbox ? msg
consumer_Recv ==
    LET result == rcv(chan_consumer_inbox) IN
    /\ can_recv(chan_consumer_inbox)                    \* Guard: channel not empty
    /\ chan_consumer_inbox' = result.channel            \* consumer.inbox ? msg
    /\ consumer_count' = consumer_count + 1             \* Update: count' = count + 1
    /\ UNCHANGED <<producer_count, step_count>>

\* Example: Chance node with probabilistic choice
\* Dice roll on every attempt to run the scheduler. R1, R2, etc
\* alter state for different predicates to activate
WorkerWithFailure ==
    /\ producer_count > 0
    /\ (
        \* choice(0 <= R1 < 90, success_path, condition, action) - 90% success
        \/ choice(0, 90, TRUE, UNCHANGED vars)
        
        \* choice(90 <= R1 < 100, failure_path, condition, action) - 10% failure
        \/ choice(90, 100, TRUE, /\ producer_count' = producer_count - 1
                                   /\ UNCHANGED <<consumer_count, chan_consumer_inbox, step_count>>)
       )

Next ==
    /\ step_count < MaxSteps
    /\ (  producer_Send
       \/ consumer_Recv
       )
    /\ step_count' = step_count + 1

\* Temporal specification
Spec == Init /\ [][Next]_vars

\* Safety properties
Safety ==
    /\ TypeOK
    /\ producer_count <= MaxMessages
    /\ Len(chan_consumer_inbox) <= Cap_consumer_inbox

\* Liveness properties
EventuallyComplete ==
    <>(producer_count = MaxMessages)

\* Fairness assumptions
Fairness ==
    /\ WF_vars(producer_Send)
    /\ WF_vars(consumer_Recv)

====
