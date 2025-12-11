---- MODULE WorkflowWithChance ----
EXTENDS Naturals, Sequences, KripkeLib

CONSTANTS MaxAttempts

VARIABLES 
    attempt_count,
    success_count,
    failure_count,
    state

vars == <<attempt_count, success_count, failure_count, state>>

Init ==
    /\ attempt_count = 0
    /\ success_count = 0
    /\ failure_count = 0
    /\ state = "ready"

\* Process with probabilistic outcome
\* Uses the choice operator from KripkeLib
AttemptTask ==
    /\ state = "ready"
    /\ attempt_count < MaxAttempts
    /\ (
        \* choice(0 <= R1 < 70, guard, action) - 70% success
        \/ choice(0, 70, TRUE,
                  /\ success_count' = success_count + 1
                  /\ state' = "success"
                  /\ UNCHANGED <<failure_count>>)
        
        \* choice(70 <= R1 < 100, guard, action) - 30% failure
        \/ choice(70, 100, TRUE,
                  /\ failure_count' = failure_count + 1
                  /\ state' = "failed"
                  /\ UNCHANGED <<success_count>>)
       )
    /\ attempt_count' = attempt_count + 1

\* Reset after handling outcome
Reset ==
    /\ state \in {"success", "failed"}
    /\ state' = "ready"
    /\ UNCHANGED <<attempt_count, success_count, failure_count>>

Next ==
    \/ AttemptTask
    \/ Reset

Spec == Init /\ [][Next]_vars

\* Safety: track attempts and outcomes correctly
TypeOK ==
    /\ attempt_count \in Nat
    /\ success_count \in Nat
    /\ failure_count \in Nat
    /\ state \in {"ready", "success", "failed"}
    /\ success_count + failure_count = attempt_count

\* Liveness: eventually we hit max attempts
EventuallyComplete == <>(attempt_count = MaxAttempts)

====
