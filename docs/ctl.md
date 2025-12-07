# CTL in `kripke-ctl`

This document describes how `kripke-ctl` understands **Computation Tree Logic (CTL)**
and how it is used to interpret English requirements over a **single,
problem-defined state machine**.

The goals:

- Make the semantics precise enough to be machine-checkable.
- Keep the notation small and stable so an English-to-CTL LLM can target it reliably.
- Explain how different *modal readings* (“must/may”, “possibly/necessarily”,
  “always/eventually”) are all just CTL formulae over one transition system.

---

## 1. Underlying model

All formulas are interpreted over a single labeled transition system

M = (S, R, L)

- **States** `S`: each state is a snapshot of all actors, channels,
  and any extra variables the model exposes.
- **Transitions** `R ⊆ S × S`: each edge is one atomic `Step`
  chosen by the scheduler.
- **Labeling** `L : S → 2^AP`: each state is labeled with the set
  of atomic propositions true in that state.

In `kripke-ctl`:

- States are produced by the actor engine (`World`, `Process`, `Step`).
- Transitions come from enabled `Step`s chosen by the scheduler.
- Propositions are boolean predicates over a state, implemented as Go functions
  and registered under string names (e.g. `Delivered`, `QueueEmpty`,
  `believes_user_logged_in`).

There is **only one** transition relation `R`.
We do **not** introduce epistemic, deontic, or counterfactual relations.
All modality is interpreted as **quantification over execution paths**
in this single structure.

---

## 2. CTL operators

We use the standard CTL operators:

- Boolean: ¬, ∧, ∨, →
- Path quantifiers: A (for all paths), E (there exists a path)
- Temporal operators:
  - X (neXt): AX, EX
  - F (Future / eventually): AF, EF
  - G (Globally / always): AG, EG
  - U (Until): EU

The four compound operators

EF, AF, EG, AG

are the core of the logic and are sufficient for all intended uses.

---

## 3. Path-based semantics

A **path** from a state s₀ is an infinite sequence

s₀, s₁, s₂, …

with (sᵢ, sᵢ₊₁) ∈ R for all i.

The satisfaction relation is standard CTL:

- EF p — there exists a path where p becomes true at some future point
- AF p — on every path, p eventually becomes true
- EG p — there exists a path where p remains true forever
- AG p — on every path, p remains true forever

---

## 4. Modal readings

All modal language is interpreted temporally over paths.

### Temporal

- “always p” → AG p
- “eventually p” → AF p
- “sometimes p” → EF p

### Alethic (temporalized)

- “possibly p (in the future)” → EF p
- “necessarily p (in the future)” → AG p

Dualities:

- EF p ≡ ¬AG ¬p
- AF p ≡ ¬EG ¬p

### Deontic (permissions and obligations)

Permissions and obligations are interpreted as properties of executions.

- **required(p)** → AG p
- **required eventually(p)** → AF p
- **allowed(p)** → EF p
- **forbidden(p)** → AG ¬p

This induces the usual deontic duality:

allowed(p) ≡ ¬ required(¬p)

No special “permission logic” is added; this is standard CTL.

---

## 5. Belief and knowledge

Belief and knowledge are **not modal operators**.

They are modeled as **state predicates**, for example:

- believes_user_logged_in
- knows_admin_secret

CTL formulas reason *about* these predicates over time:

AG (believes_user_logged_in → AF corrected)
AG (knows_admin_secret → AG protected)

Beliefs may be false, inconsistent, or outdated.
No epistemic axioms are assumed.

---

## 6. Summary

- One state machine
- One transition relation
- One CTL semantics
- EF / AF / EG / AG as the only modal operators
- Deontic and alethic readings are interpretations, not new logics
- Belief and knowledge live in state

This structure is intentionally small, explicit, and stable,
so that English requirements can be compiled reliably into CTL
and checked mechanically against the model.
