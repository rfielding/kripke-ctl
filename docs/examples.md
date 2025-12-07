ENGLISH TO MODEL EXAMPLES
========================

This file gives examples of English requirements and their corresponding
CTL formulas, using the semantics defined in ctl.md and engine.md.

Conventions:
- All predicates are boolean predicates over World.
- CTL is evaluated over a single Kripke structure produced by the engine.
- “must / required”  → AG
- “may / allowed”    → EF
- “eventually”       → AF
- “believes / knows” → state predicates (not modal operators)

-----------------------------------------------------------------------

1. SAFETY (INVARIANT)

English:
  The system must never lose a message.

State predicates:
  LostMessage

CTL:
  AG ¬LostMessage

Meaning:
  On all paths, in all future states, no message is lost.

-----------------------------------------------------------------------

2. LIVENESS (EVENTUAL COMPLETION)

English:
  Every order is eventually delivered.

State predicates:
  PendingOrder

CTL:
  AF ¬PendingOrder

Meaning:
  On every execution, there is a future state where no orders remain pending.

-----------------------------------------------------------------------

3. PERMISSION (ALLOWED BEHAVIOR)

English:
  The system may retry a failed payment.

State predicates:
  RetryPayment

CTL:
  EF RetryPayment

Meaning:
  There exists some execution where a retry occurs.

-----------------------------------------------------------------------

4. OBLIGATION VS PERMISSION

English:
  The system must retry a failed payment.

State predicates:
  FailedPayment
  RetryPayment

CTL:
  AG (FailedPayment → AF RetryPayment)

Meaning:
  On every path, whenever a payment fails, a retry eventually happens.

-----------------------------------------------------------------------

5. FORBIDDEN BEHAVIOR

English:
  The system must never process an order without authorization.

State predicates:
  UnauthorizedOrderProcessed

CTL:
  AG ¬UnauthorizedOrderProcessed

-----------------------------------------------------------------------

6. LOOPHOLE / COUNTEREXAMPLE

English:
  It is possible for the system to run forever without delivering an order.

State predicates:
  OrderDelivered

CTL:
  EG ¬OrderDelivered

Meaning:
  There exists an infinite execution where delivery never occurs.

-----------------------------------------------------------------------

7. ABILITY VS GUARANTEE

English:
  The system can recover from an error.

State predicates:
  Error
  Recovered

CTL:
  AG (Error → EF Recovered)

Meaning:
  From any error state, recovery is possible.

-----------------------------------------------------------------------

8. GUARANTEED RECOVERY

English:
  The system always recovers from an error.

CTL:
  AG (Error → AF Recovered)

-----------------------------------------------------------------------

9. BELIEF AS STATE (NO EPISTEMIC MODALITY)

English:
  If the user believes they are logged in, they will eventually be corrected
  if that belief is false.

State predicates:
  BelievesLoggedIn
  LoggedIn
  Corrected

CTL:
  AG ((BelievesLoggedIn ∧ ¬LoggedIn) → AF Corrected)

Meaning:
  Belief is ordinary state; CTL reasons about how it evolves.

-----------------------------------------------------------------------

10. KNOWLEDGE IMPLIES SAFETY

English:
  If an admin knows the secret, it is always protected.

State predicates:
  KnowsSecret
  SecretProtected

CTL:
  AG (KnowsSecret → AG SecretProtected)

-----------------------------------------------------------------------

11. CHANNEL SEMANTICS

English:
  Every message placed in a queue is eventually dequeued.

State predicates:
  InQueue(msg)
  Dequeued(msg)

CTL:
  AG (InQueue(msg) → AF Dequeued(msg))

(Note: quantification over msg is handled by model construction.)

-----------------------------------------------------------------------

12. FAIRNESS AS A REQUIREMENT

English:
  If a process is continuously enabled, it will eventually run.

State predicates:
  Enabled(p)
  Ran(p)

CTL:
  AG (Enabled(p) → AF Ran(p))

Meaning:
  Fairness is expressed as a property, not an assumption.

-----------------------------------------------------------------------

13. ERROR REACHABILITY (DEBUGGING)

English:
  Is it possible for the system to reach an error state?

State predicates:
  Error

CTL:
  EF Error

-----------------------------------------------------------------------

14. DEONTIC + BELIEF

English:
  The system must never act on a belief known to be false.

State predicates:
  Believes(p)
  ¬p
  ActedOn(p)

CTL:
  AG ((Believes(p) ∧ ¬p) → ¬ActedOn(p))

-----------------------------------------------------------------------

15. SUMMARY HEURISTIC FOR LLMs

English phrase → CTL form

  must / required        → AG
  always                 → AG
  eventually             → AF
  may / allowed / can    → EF
  never p                → AG ¬p
  can avoid forever p    → EG ¬p
  believes / knows       → state predicate

These examples are intended to be directly compiled by an LLM into
CTL formulas and then model-checked against the engine’s state graph.
