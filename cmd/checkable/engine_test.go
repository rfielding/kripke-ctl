; demo_requirements.lisp
;
; Assumes: prologue.lisp and kripke.lisp already loaded.
;
; Goal:
;   - Provide a runnable "requirements negotiation" loop skeleton
;   - Generate a Markdown spec with Mermaid diagrams + CTL proof results
;   - Exercise your bounded/scheduled execution model (actors + mailboxes)
;
; What works immediately:
;   - Actor scheduler exercise (spawn-actor, send-to!, receive!, run-scheduler)
;   - Markdown generation stored in registry
;
; What you must adapt to your kripke.lisp API:
;   - engine:build-model
;   - engine:prove
;   - engine:render-mermaid


; -----------------------------------------------------------------------------
; Small utilities (no dependency on extra libs)
; -----------------------------------------------------------------------------

(define (nl) "\n")

(define (str . xs)
  (apply string-append (map repr xs))) ; repr returns strings in your runtime

(define (md-h1 s) (string-append "# " s (nl) (nl)))
(define (md-h2 s) (string-append "## " s (nl) (nl)))
(define (md-code fence body)
  (string-append "```" fence (nl) body (nl) "```" (nl) (nl)))

(define (md-bullets items)
  (define (go items acc)
    (if (empty? items)
        acc
        (go (rest items)
            (string-append acc "- " (first items) (nl)))))
  (string-append (go items "") (nl)))

(define (md-quote s)
  (string-append "> " s (nl) (nl)))


; -----------------------------------------------------------------------------
; Engine adapter layer (YOU wire these to kripke.lisp exports)
; -----------------------------------------------------------------------------
;
; The idea is: keep the tutorial/demo stable, and only change these adapters
; to match whatever functions exist in kripke.lisp.
;

(define (engine:build-model reqs)
  ; TODO: replace with your real model builder.
  ; Should return an opaque "model" value used by engine:prove and engine:render-mermaid.
  ;
  ; Example idea (not real): (kripke:from-requirements reqs)
  (tag 'mock-model reqs))

(define (engine:prove model ctl-expr)
  ; TODO: replace with your real prover/checker.
  ; Should return something like: true/false (or a richer structure).
  ;
  ; Example idea (not real): (ctl:check model ctl-expr)
  (tag 'mock-proof (list model ctl-expr true)))

(define (engine:render-mermaid model)
  ; TODO: replace with your real diagram generator.
  ; Should return a Mermaid string.
  ;
  ; Example idea (not real): (kripke:mermaid model)
  (string-append
    "stateDiagram-v2\n"
    "  [*] --> Draft\n"
    "  Draft --> Refined: negotiate\n"
    "  Refined --> Proved: prove\n"
    "  Proved --> Refined: counterexample\n"
    "  Proved --> [*]\n"))


; -----------------------------------------------------------------------------
; Spec state (stored in registry so the whole session can append)
; -----------------------------------------------------------------------------

(define (spec:init!)
  (registry-set! "spec:title" "Kripke-CTL Requirements Session")
  (registry-set! "spec:reqs" (list))           ; list of requirement strings
  (registry-set! "spec:ctl"  (list))           ; list of ctl-expr strings (or exprs)
  (registry-set! "spec:notes" (list))          ; freeform notes
  (registry-set! "spec:md" "")                 ; generated markdown
  'ok)

(define (spec:add-req! s)
  (registry-set! "spec:reqs" (append (registry-get "spec:reqs") (list s)))
  'ok)

(define (spec:add-ctl! ctl-expr)
  ; ctl-expr can be a string or a data structure, your choice.
  (registry-set! "spec:ctl" (append (registry-get "spec:ctl") (list ctl-expr)))
  'ok)

(define (spec:add-note! s)
  (registry-set! "spec:notes" (append (registry-get "spec:notes") (list s)))
  'ok)


; -----------------------------------------------------------------------------
; Markdown generator
; -----------------------------------------------------------------------------

(define (spec:render-md!)
  (define title (registry-get "spec:title"))
  (define reqs  (registry-get "spec:reqs"))
  (define ctls  (registry-get "spec:ctl"))
  (define notes (registry-get "spec:notes"))

  ; Build a model from requirements (adapter)
  (define model (engine:build-model reqs))

  ; Render Mermaid from model (adapter)
  (define mermaid (engine:render-mermaid model))

  ; Prove each CTL predicate (adapter)
  (define (prove-all ctls acc)
    (if (empty? ctls)
        acc
        (let proof (engine:prove model (first ctls))
          (prove-all (rest ctls) (append acc (list proof))))))
  (define proofs (prove-all ctls (list)))

  ; Turn requirements into markdown bullet strings
  (define (stringify xs)
    (define (go xs acc)
      (if (empty? xs)
          acc
          (go (rest xs) (append acc (list (first xs))))))
    (go xs (list)))

  ; proofs -> readable lines (placeholder; you can pretty-print later)
  (define (proof-lines ps acc)
    (if (empty? ps)
        acc
        (proof-lines (rest ps)
                     (append acc (list (repr (first ps)))))))

  (define md
    (string-append
      (md-h1 title)

      (md-h2 "Intent")
      (md-quote "Negotiate requirements with an LLM, compile them into a Kripke model, render diagrams, and check CTL predicates.")

      (md-h2 "Requirements")
      (md-bullets (stringify reqs))

      (md-h2 "Notes")
      (md-bullets (stringify notes))

      (md-h2 "Model")
      (md-code "mermaid" mermaid)

      (md-h2 "CTL Proofs")
      (md-code "" (apply string-append (map (lambda (s) (string-append "- " s (nl)))
                                            (proof-lines proofs (list)))))

      (md-h2 "Repro")
      (md-code ""
        (string-append
          "; This file assumes prologue.lisp + kripke.lisp already loaded.\n"
          "; Run: (demo:run)\n"
          "; Then read: (registry-get \"spec:md\")\n"))
    ))

  (registry-set! "spec:md" md)
  md)


; -----------------------------------------------------------------------------
; A concrete scheduler exercise: two actors negotiating "spec deltas"
; -----------------------------------------------------------------------------
;
; We simulate the “LLM negotiation” using two actors:
;   - coordinator: applies deltas, rebuilds markdown, and stores it in registry
;   - mock-llm: sends a few requirement + ctl messages, then done
;
; Later you can replace mock-llm with a real integration that posts messages
; into the coordinator’s mailbox.
;

; Message protocol (as data):
;   (req "text...")
;   (ctl <expr-or-string>)
;   (note "text...")
;   (render)
;   (done)

(define (demo:coordinator-loop)
  (let msg (receive!)
    (match msg
      ((list 'req ?s)
        (begin
          (spec:add-req! ?s)
          (list 'continue (list 'demo:coordinator-loop))))

      ((list 'ctl ?e)
        (begin
          (spec:add-ctl! ?e)
          (list 'continue (list 'demo:coordinator-loop))))

      ((list 'note ?s)
        (begin
          (spec:add-note! ?s)
          (list 'continue (list 'demo:coordinator-loop))))

      ((list 'render)
        (begin
          (spec:render-md!)
          (list 'continue (list 'demo:coordinator-loop))))

      ((list 'done)
        (begin
          (spec:render-md!)
          (done!)))

      (_
        (begin
          (spec:add-note! (string-append "Unknown message: " (repr msg)))
          (list 'continue (list 'demo:coordinator-loop)))))))

(define (demo:mock-llm-script)
  (begin
    ; Send a tiny tutorial-ish set of requirements
    (send-to! 'coordinator (list 'req "All variable-length structures have explicit capacities (bounded stacks/queues)."))
    (send-to! 'coordinator (list 'req "Actors expose input-only mailboxes; they communicate by sending messages, never sharing mutable state."))
    (send-to! 'coordinator (list 'req "Blocking is modeled as scheduler-unschedulable (actor becomes blocked on a resource condition)."))
    (send-to! 'coordinator (list 'note "We will iterate by sending deltas; coordinator regenerates Markdown after each batch."))
    (send-to! 'coordinator (list 'note "Hook this to a real LLM by injecting messages into coordinator’s mailbox."))

    ; CTL examples (as strings for now)
    (send-to! 'coordinator (list 'ctl "AG(not deadlocked)"))
    (send-to! 'coordinator (list 'ctl "EF(completed)"))

    ; Ask coordinator to render mid-way
    (send-to! 'coordinator (list 'render))

    ; Add more requirements
    (send-to! 'coordinator (list 'req "State selection is guard-based; transitions are explicit actions. Probabilities live on actions, not states."))
    (send-to! 'coordinator (list 'ctl "AG(queue_len <= capacity)"))

    ; Finish
    (send-to! 'coordinator (list 'done))
    (done!)))


; -----------------------------------------------------------------------------
; Entry point
; -----------------------------------------------------------------------------

(define (demo:run)
  (begin
    (spec:init!)
    (set-trace! false)

    (spawn-actor 'coordinator 64 (list 'demo:coordinator-loop))
    (spawn-actor 'mock-llm     64 (list 'demo:mock-llm-script))

    (run-scheduler 10000)

    ; Return the markdown so you see something in the REPL/file-run
    (registry-get "spec:md")))

