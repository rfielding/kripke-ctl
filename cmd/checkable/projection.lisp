; ============================================================================
; Projection System
; Derives local actor behaviors from global grammar choreographies
; ============================================================================

; ----------------------------------------------------------------------------
; Grammar Analysis Helpers
; ----------------------------------------------------------------------------

; A transition looks like: (Sender -> Receiver : msg-type ?bindings...) -> NextState
; We need to parse these from grammar state definitions

; Check if something is a transition arrow
(define (arrow? x)
  (if (symbol? x)
      (= x '->)
      false))

; Check if something is a message separator
(define (colon? x)
  (if (symbol? x)
      (= x ':)
      false))

; Drop first n elements from a list
(define (drop lst n)
  (if (or (empty? lst) (<= n 0))
      lst
      (drop (rest lst) (- n 1))))

; Parse a single transition from a state body
; Input: ((Sender -> Receiver : msg args...) -> NextState rest...)
; Output: ((sender Sender) (receiver Receiver) (msg-type msg) (args (args...)) (next NextState))
(define (parse-transition items)
  (if (empty? items)
      nil
      (let first-item (first items)
        (if (list? first-item)
            ; It's a message spec: (Sender -> Receiver : msg args...)
            (let spec first-item
              (let sender (nth spec 0)
                (let receiver (nth spec 2)  ; skip ->
                  (let msg-type (nth spec 4)  ; skip :
                    (let args (drop spec 5)
                      (let rest-items (rest items)
                        (if (and (not (empty? rest-items)) (arrow? (first rest-items)))
                            (let next-state (nth rest-items 1)
                              (list
                                (list 'sender sender)
                                (list 'receiver receiver)
                                (list 'msg-type msg-type)
                                (list 'args args)
                                (list 'next next-state)
                                (list 'remaining (drop rest-items 2))))
                            nil)))))))
            nil))))

; Get all transitions from a state definition
; A state looks like: (StateName trans1 -> Next1 trans2 -> Next2 ...)
(define (get-state-transitions state-def)
  (let state-name (first state-def)
    (let body (rest state-def)
      (parse-all-transitions state-name body '()))))

(define (parse-all-transitions state-name items acc)
  (if (empty? items)
      acc
      (let parsed (parse-transition items)
        (if (nil? parsed)
            acc
            (let trans (list
                        (list 'from state-name)
                        (list 'sender (second (first parsed)))
                        (list 'receiver (second (nth parsed 1)))
                        (list 'msg-type (second (nth parsed 2)))
                        (list 'args (second (nth parsed 3)))
                        (list 'next (second (nth parsed 4))))
              (let remaining (second (nth parsed 5))
                (parse-all-transitions state-name remaining (append acc (list trans)))))))))

; Get a field from a transition association list
(define (trans-get trans field)
  (assoc-get trans field))

(define (assoc-get alist key)
  (if (empty? alist)
      nil
      (let pair (first alist)
        (if (list? pair)
            (if (= (first pair) key)
                (if (> (length pair) 1)
                    (second pair)
                    nil)
                (assoc-get (rest alist) key))
            (assoc-get (rest alist) key)))))

; Get second element
(define (second lst)
  (nth lst 1))

; ----------------------------------------------------------------------------
; Role Extraction
; ----------------------------------------------------------------------------

; Get all roles mentioned in a grammar
(define (extract-roles grammar-data)
  (let states (rest grammar-data)  ; skip grammar name
    (extract-roles-from-states states '())))

(define (extract-roles-from-states states acc)
  (if (empty? states)
      (unique acc)
      (let state (first states)
        (let transitions (get-state-transitions state)
          (let roles (extract-roles-from-transitions transitions)
            (extract-roles-from-states (rest states) (append acc roles)))))))

(define (extract-roles-from-transitions transitions)
  (if (empty? transitions)
      '()
      (let trans (first transitions)
        (let sender (trans-get trans 'sender)
          (let receiver (trans-get trans 'receiver)
            (append (list sender receiver)
                    (extract-roles-from-transitions (rest transitions))))))))

; Remove duplicates from a list
(define (unique lst)
  (unique-helper lst '()))

(define (unique-helper lst acc)
  (if (empty? lst)
      acc
      (let item (first lst)
        (if (member? item acc)
            (unique-helper (rest lst) acc)
            (unique-helper (rest lst) (append acc (list item)))))))

(define (member? item lst)
  (if (empty? lst)
      false
      (if (= item (first lst))
          true
          (member? item (rest lst)))))

; ----------------------------------------------------------------------------
; Projection
; ----------------------------------------------------------------------------

; Project grammar to a specific role's view
; Returns list of (local-state transitions) where transitions are role-relevant
(define (project-grammar grammar-data role)
  (let states (rest grammar-data)
    (project-states states role)))

(define (project-states states role)
  (if (empty? states)
      '()
      (let state (first states)
        (let state-name (first state)
          (let transitions (get-state-transitions state)
            (let my-transitions (filter-transitions-for-role transitions role)
              (cons (list state-name my-transitions)
                    (project-states (rest states) role))))))))

; Filter transitions to only those involving this role
(define (filter-transitions-for-role transitions role)
  (if (empty? transitions)
      '()
      (let trans (first transitions)
        (let sender (trans-get trans 'sender)
          (let receiver (trans-get trans 'receiver)
            (if (or (= sender role) (= receiver role))
                (cons trans (filter-transitions-for-role (rest transitions) role))
                (filter-transitions-for-role (rest transitions) role)))))))

; ----------------------------------------------------------------------------
; Code Generation
; ----------------------------------------------------------------------------

; Generate actor behavior functions from a projection
; Each state becomes a function that:
; - If I'm sender: send message, become next state
; - If I'm receiver: wait for message, become next state
; - If multiple transitions: check which one matches

(define (generate-actor-code role projection role-bindings)
  (let state-fns (map-states role projection role-bindings)
    state-fns))

(define (map-states role states role-bindings)
  (if (empty? states)
      '()
      (let state (first states)
        (let state-name (first state)
          (let transitions (second state)
            (let fn-name (make-state-fn-name role state-name)
              (let fn-body (generate-state-body role transitions role-bindings)
                (cons (list 'define (list fn-name) fn-body)
                      (map-states role (rest states) role-bindings)))))))))

(define (make-state-fn-name role state-name)
  ; Convert to symbol like: acquirer-Start
  (string->symbol (string-append (symbol->string role) "-" (symbol->string state-name))))

(define (generate-state-body role transitions role-bindings)
  (if (empty? transitions)
      ; No transitions for me in this state - I'm done or it's a terminal
      ''done
      (if (= (length transitions) 1)
          ; Single transition
          (generate-single-transition role (first transitions) role-bindings)
          ; Multiple transitions - need to branch
          (generate-multi-transition role transitions role-bindings))))

(define (generate-single-transition role trans role-bindings)
  (let sender (trans-get trans 'sender)
    (let receiver (trans-get trans 'receiver)
      (let msg-type (trans-get trans 'msg-type)
        (let next-state (trans-get trans 'next)
          (let receiver-actor (lookup-role receiver role-bindings)
            (let next-fn (make-state-fn-name role next-state)
              (if (= sender role)
                  ; I'm sending
                  (list 'do
                    (list 'send-to! (list 'quote receiver-actor) (list 'quote (list msg-type)))
                    (list 'list (list 'quote 'become) (list 'quote (list next-fn))))
                  ; I'm receiving
                  (list 'let 'msg '(receive-now!)
                    (list 'if '(= msg 'empty)
                      ''yield
                      (list 'do
                        (list 'list (list 'quote 'become) (list 'quote (list next-fn))))))))))))))

(define (generate-multi-transition role transitions role-bindings)
  ; For multiple possible transitions, generate a cond/match
  ; If I'm receiving, match on message type
  ; If I'm sending, this is a choice point (non-deterministic or guided by some condition)
  (list 'let 'msg '(receive-now!)
    (list 'if '(= msg 'empty)
      ''yield
      (generate-match-cases role transitions role-bindings))))

(define (generate-match-cases role transitions role-bindings)
  (if (empty? transitions)
      ''yield  ; no match, yield
      (let trans (first transitions)
        (let msg-type (trans-get trans 'msg-type)
          (let next-state (trans-get trans 'next)
            (let next-fn (make-state-fn-name role next-state)
              (list 'if (list '= '(first msg) (list 'quote msg-type))
                (list 'list ''become (list 'quote (list next-fn)))
                (generate-match-cases role (rest transitions) role-bindings))))))))

(define (lookup-role role bindings)
  (assoc-get bindings role))

; ----------------------------------------------------------------------------
; High-Level API
; ----------------------------------------------------------------------------

; Create actors from a grammar with role->actor-name bindings
; Example: (instantiate-grammar 'acquisition '((Acquirer . company-a) (Target . company-b) (Regulator . ftc)))
(define (instantiate-grammar grammar-name bindings)
  (let grammar-data (get-grammar grammar-name)
    (if (nil? grammar-data)
        (do (println "Grammar not found:" grammar-name) nil)
        (let roles (extract-roles grammar-data)
          (do
            (println "Roles:" roles)
            (instantiate-roles grammar-data roles bindings))))))

(define (instantiate-roles grammar-data roles bindings)
  (if (empty? roles)
      'ok
      (let role (first roles)
        (let actor-name (lookup-role role bindings)
          (if (nil? actor-name)
              (do 
                (println "No binding for role:" role)
                (instantiate-roles grammar-data (rest roles) bindings))
              (do
                (println "Creating actor" actor-name "for role" role)
                (let projection (project-grammar grammar-data role)
                  (do
                    ; Generate state functions for this role
                    (generate-and-register-states role actor-name projection bindings)
                    ; Spawn the actor with initial state
                    (let initial-state (find-initial-state projection role)
                      (do
                        (println "  Initial state:" initial-state)
                        (spawn-actor actor-name 16 initial-state)
                        (instantiate-roles grammar-data (rest roles) bindings)))))))))))

; Find the first state where this role has something to do
(define (find-initial-state projection role)
  (find-initial-state-helper projection role))

(define (find-initial-state-helper states role)
  (if (empty? states)
      ''done  ; No states with transitions
      (let state (first states)
        (let state-name (first state)
          (let transitions (second state)
            (if (empty? transitions)
                (find-initial-state-helper (rest states) role)
                ; Found a state with transitions - generate call to state function
                (let fn-name (make-state-fn-name role state-name)
                  (list fn-name))))))))

; Generate state functions and register them
(define (generate-and-register-states role actor-name projection bindings)
  ; Store projection globally for find-next-active-state
  (set! *current-projection* projection)
  (generate-states-helper role actor-name projection bindings))

(define (generate-states-helper role actor-name states bindings)
  (if (empty? states)
      'ok
      (let state (first states)
        (let state-name (first state)
          (let transitions (second state)
            (do
              (generate-state-function role actor-name state-name transitions bindings)
              (generate-states-helper role actor-name (rest states) bindings)))))))

; Generate a single state function
(define (generate-state-function role actor-name state-name transitions bindings)
  (let fn-name (make-state-fn-name role state-name)
    (if (empty? transitions)
        ; No transitions for me in this state
        (let next-active (find-next-active-state role state-name *current-projection*)
          (if (= next-active state-name)
              ; No next active found - this is terminal for this role
              (eval (list 'define (list fn-name) ''done))
              ; Transition to next active state
              (let next-fn (make-state-fn-name role next-active)
                (eval (list 'define (list fn-name)
                       (list 'list ''become (list 'quote (list next-fn))))))))
        ; State with transitions
        (let body (generate-state-body-from-transitions role transitions bindings)
          (eval (list 'define (list fn-name) body))))))

; Generate the body of a state function from its transitions
(define (generate-state-body-from-transitions role transitions bindings)
  (let am-sender (all-sender? role transitions)
    (let am-receiver (all-receiver? role transitions)
      (if am-sender
          ; All transitions have me as sender - I choose and send
          (generate-sender-choice role transitions bindings)
          (if am-receiver
              ; All transitions have me as receiver - I wait and match
              (generate-receiver-match role transitions bindings)
              ; Mixed - shouldn't happen in well-formed choreography
              (generate-receiver-match role transitions bindings))))))

; Check if role is sender in all transitions
(define (all-sender? role transitions)
  (if (empty? transitions)
      true
      (let trans (first transitions)
        (if (= (trans-get trans 'sender) role)
            (all-sender? role (rest transitions))
            false))))

; Check if role is receiver in all transitions
(define (all-receiver? role transitions)
  (if (empty? transitions)
      true
      (let trans (first transitions)
        (if (= (trans-get trans 'receiver) role)
            (all-receiver? role (rest transitions))
            false))))

; Generate code for sender with choice - for now, pick first option
; (In a real system, this would be guided by user policy or non-determinism)
(define (generate-sender-choice role transitions bindings)
  (if (empty? transitions)
      ''done
      ; Pick first transition as default choice
      (let trans (first transitions)
        (generate-single-trans-body role trans bindings))))

; Generate code for receiver with multiple possible incoming messages
(define (generate-receiver-match role transitions bindings)
  (if (= (length transitions) 1)
      (generate-single-trans-body role (first transitions) bindings)
      (list 'let 'msg '(receive-now!)
        (list 'if '(= msg 'empty)
          ''yield
          (generate-match-chain role transitions bindings)))))

; Single transition: either send or receive
; When transitioning, skip to the next state where this role has action
(define (generate-single-trans-body role trans bindings)
  (let sender (trans-get trans 'sender)
    (let receiver (trans-get trans 'receiver)
      (let msg-type (trans-get trans 'msg-type)
        (let grammar-next (trans-get trans 'next)
          ; Find next state where role has transitions
          (let next-state (find-next-active-state role grammar-next *current-projection*)
            (let next-fn (make-state-fn-name role next-state)
              (if (= sender role)
                  ; I'm sending
                  (let target-actor (lookup-role receiver bindings)
                    (list 'do
                      (list 'send-to! (list 'quote target-actor) (list 'quote (list msg-type)))
                      (list 'list ''become (list 'quote (list next-fn)))))
                  ; I'm receiving
                  (list 'let 'msg '(receive-now!)
                    (list 'if '(= msg 'empty)
                      ''yield
                      (list 'if (list '= '(first msg) (list 'quote msg-type))
                        (list 'list ''become (list 'quote (list next-fn)))
                        ''yield)))))))))))

; Generated function for states with no transitions should immediately
; become the next state where this role IS active (has transitions)
; This requires computing the "skip map" when generating code

; Global to hold current projection during code generation
(define *current-projection* nil)

; Compute the next active state for a role starting from state-name
; Returns the first state in projection order where role has transitions
(define (find-next-active-state role state-name projection)
  (find-active-from role state-name projection))

(define (find-active-from role start-state projection)
  ; Find start-state in projection, then search forward
  (let found-start (find-index-of start-state projection 0)
    (if (< found-start 0)
        start-state  ; Not found, return as-is
        ; Search from found-start onward for a state with transitions
        (find-first-active-from found-start projection start-state))))

(define (find-index-of state-name projection idx)
  (if (empty? projection)
      -1
      (if (= (first (first projection)) state-name)
          idx
          (find-index-of state-name (rest projection) (+ idx 1)))))

(define (find-first-active-from idx projection default)
  (let remaining (drop-n projection idx)
    (if (empty? remaining)
        default  ; No more states, return default (which is the start state)
        (let entry (first remaining)
          (let state-name (first entry)
            (let transitions (second entry)
              (if (not (empty? transitions))
                  state-name  ; Has transitions - this is an active state
                  ; No transitions - check if there are more states after this
                  (let next-remaining (rest remaining)
                    (if (empty? next-remaining)
                        state-name  ; Last state, it's terminal
                        ; More states exist, keep searching
                        (find-first-active-from (+ idx 1) projection default))))))))))

(define (drop-n lst n)
  (if (or (<= n 0) (empty? lst))
      lst
      (drop-n (rest lst) (- n 1))))

(define (find-state-in-projection state-name projection)
  (if (empty? projection)
      nil
      (let entry (first projection)
        (if (= (first entry) state-name)
            entry
            (find-state-in-projection state-name (rest projection))))))

; Generate a chain of if statements to match message types
(define (generate-match-chain role transitions bindings)
  (if (empty? transitions)
      ''yield
      (let trans (first transitions)
        (let sender (trans-get trans 'sender)
          (let msg-type (trans-get trans 'msg-type)
            (let next-state (trans-get trans 'next)
              (let next-fn (make-state-fn-name role next-state)
                (if (= sender role)
                    ; I'm sender in this branch - this is a choice point
                    ; For now, skip sender branches in receive context
                    (generate-match-chain role (rest transitions) bindings)
                    ; I'm receiver - match on message type
                    (list 'if (list '= '(first msg) (list 'quote msg-type))
                      (list 'list ''become (list 'quote (list next-fn)))
                      (generate-match-chain role (rest transitions) bindings))))))))))

(define (map fn lst)
  (if (empty? lst)
      '()
      (cons (fn (first lst)) (map fn (rest lst)))))

; ----------------------------------------------------------------------------
; Debug / Inspection
; ----------------------------------------------------------------------------

(define (show-projection grammar-name role)
  (let grammar-data (get-grammar grammar-name)
    (if (nil? grammar-data)
        (println "Grammar not found")
        (let projection (project-grammar grammar-data role)
          (show-projection-states projection)))))

(define (show-projection-states states)
  (if (empty? states)
      nil
      (let state (first states)
        (let state-name (first state)
          (let transitions (second state)
            (do
              (println "\nState:" state-name)
              (show-transitions transitions)
              (show-projection-states (rest states))))))))

(define (show-transitions transitions)
  (if (empty? transitions)
      nil
      (let trans (first transitions)
        (do
          (println "  " (trans-get trans 'sender) "->" 
                       (trans-get trans 'receiver) ":"
                       (trans-get trans 'msg-type) "->"
                       (trans-get trans 'next))
          (show-transitions (rest transitions))))))

; ----------------------------------------------------------------------------
(println "Projection system loaded.")
