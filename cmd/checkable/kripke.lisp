; ============================================================================
; Kripke Structure Extraction
; Builds state machines from grammar definitions for CTL model checking
; ============================================================================

; ----------------------------------------------------------------------------
; Helper Functions
; ----------------------------------------------------------------------------

(define (map f lst)
  (if (empty? lst)
      '()
      (cons (f (first lst)) (map f (rest lst)))))

(define (second lst)
  (nth lst 1))

(define (drop lst n)
  (if (or (empty? lst) (<= n 0))
      lst
      (drop (rest lst) (- n 1))))

; Filter helper
(define (filter pred lst)
  (if (empty? lst)
      '()
      (if (pred (first lst))
          (cons (first lst) (filter pred (rest lst)))
          (filter pred (rest lst)))))

; Helper: check if any element satisfies predicate
(define (any? pred lst)
  (if (empty? lst)
      false
      (if (pred (first lst))
          true
          (any? pred (rest lst)))))

; Helper: check if all elements satisfy predicate
(define (all? pred lst)
  (if (empty? lst)
      true
      (if (pred (first lst))
          (all? pred (rest lst))
          false)))

; ----------------------------------------------------------------------------
; Grammar Analysis
; ----------------------------------------------------------------------------

; A Kripke structure is: (states initial-states transitions labels)
; where:
;   states = list of state names
;   initial-states = list of initial state names  
;   transitions = list of (from to label) triples
;   labels = list of (state . propositions) pairs

; Extract all state names from a grammar
(define (grammar-states grammar-data)
  (if (nil? grammar-data)
      '()
      (map (lambda (s) (first s)) (rest grammar-data))))

; Extract all transitions from a grammar
; Returns list of (from-state to-state sender receiver msg-type)
(define (grammar-transitions grammar-data)
  (if (nil? grammar-data)
      '()
      (collect-all-transitions (rest grammar-data))))

(define (collect-all-transitions states)
  (if (empty? states)
      '()
      (let state-def (first states)
        (let state-name (first state-def)
          (let trans (extract-transitions-from-state state-def)
            (append trans (collect-all-transitions (rest states))))))))

; Extract transitions from a single state definition
; State def: (StateName (Sender -> Receiver : msg) -> NextState ...)
(define (extract-transitions-from-state state-def)
  (let state-name (first state-def)
    (extract-trans-from-body state-name (rest state-def) '())))

(define (extract-trans-from-body from-state body acc)
  (if (empty? body)
      acc
      (let item (first body)
        (if (list? item)
            ; It's a message spec (Sender -> Receiver : msg)
            (let rest-body (rest body)
              (if (and (not (empty? rest-body)) 
                       (symbol? (first rest-body))
                       (= (first rest-body) '->))
                  ; Next item is ->, then comes the target state
                  (let to-state (nth rest-body 1)
                    (let sender (nth item 0)
                      (let receiver (nth item 2)
                        (let msg-type (nth item 4)
                          (let trans (list from-state to-state sender receiver msg-type)
                            (extract-trans-from-body from-state (drop rest-body 2)
                                                    (append acc (list trans))))))))
                  (extract-trans-from-body from-state rest-body acc)))
            (extract-trans-from-body from-state (rest body) acc)))))

; ----------------------------------------------------------------------------
; Kripke Structure Construction  
; ----------------------------------------------------------------------------

; Build a Kripke structure from a grammar
(define (grammar->kripke grammar-name)
  (let grammar-data (get-grammar grammar-name)
    (if (nil? grammar-data)
        nil
        (let states (grammar-states grammar-data)
          (let transitions (grammar-transitions grammar-data)
            (let initial (list (first states))  ; First state is initial
              (let labels (generate-state-labels states transitions)
                (tag 'kripke
                  (list
                    (list 'name grammar-name)
                    (list 'states states)
                    (list 'initial initial)
                    (list 'transitions transitions)
                    (list 'labels labels))))))))))

; Generate atomic proposition labels for each state
; Each state gets: in_StateName proposition
; Each transition adds: sender_sends_msg, receiver_receives_msg
(define (generate-state-labels states transitions)
  (map (lambda (s) (list s (state-propositions s transitions))) states))

(define (state-propositions state-name transitions)
  ; Start with in_State proposition
  (let base-props (list (make-in-prop state-name))
    ; Add propositions for outgoing transitions
    (let outgoing (filter (lambda (t) (= (first t) state-name)) transitions)
      (append base-props (collect-transition-props outgoing)))))

(define (make-in-prop state-name)
  (string->symbol (string-append "in_" (symbol->string state-name))))

(define (collect-transition-props transitions)
  (if (empty? transitions)
      '()
      (let t (first transitions)
        (let sender (nth t 2)
          (let receiver (nth t 3)
            (let msg (nth t 4)
              (append 
                (list 
                  (make-sends-prop sender msg)
                  (make-receives-prop receiver msg))
                (collect-transition-props (rest transitions)))))))))

(define (make-sends-prop role msg)
  (string->symbol (string-append (symbol->string role) "_sends_" (symbol->string msg))))

(define (make-receives-prop role msg)
  (string->symbol (string-append (symbol->string role) "_receives_" (symbol->string msg))))

; ----------------------------------------------------------------------------
; Kripke Structure Accessors
; ----------------------------------------------------------------------------

(define (kripke? k)
  (and (not (nil? k)) (tag-is? k 'kripke)))

(define (kripke-name k)
  (if (kripke? k)
      (second (first (tag-value k)))
      nil))

(define (kripke-states k)
  (if (kripke? k)
      (second (nth (tag-value k) 1))
      '()))

(define (kripke-initial k)
  (if (kripke? k)
      (second (nth (tag-value k) 2))
      '()))

(define (kripke-transitions k)
  (if (kripke? k)
      (second (nth (tag-value k) 3))
      '()))

(define (kripke-labels k)
  (if (kripke? k)
      (second (nth (tag-value k) 4))
      '()))

; Get successor states from a given state
(define (kripke-successors k state)
  (let trans (kripke-transitions k)
    (map second (filter (lambda (t) (= (first t) state)) trans))))

; Get propositions true in a state
(define (kripke-props-at k state)
  (let labels (kripke-labels k)
    (let entry (find-label state labels)
      (if (nil? entry)
          '()
          (second entry)))))

(define (find-label state labels)
  (if (empty? labels)
      nil
      (if (= (first (first labels)) state)
          (first labels)
          (find-label state (rest labels)))))

; ----------------------------------------------------------------------------
; Display
; ----------------------------------------------------------------------------

(define (print-kripke k)
  (if (not (kripke? k))
      (println "Not a Kripke structure")
      (do
        (println "Kripke Structure:" (kripke-name k))
        (println "")
        (println "States:" (kripke-states k))
        (println "Initial:" (kripke-initial k))
        (println "")
        (println "Transitions:")
        (print-transitions (kripke-transitions k))
        (println "")
        (println "Labels:")
        (print-labels (kripke-labels k)))))

(define (print-transitions trans)
  (if (empty? trans)
      nil
      (let t (first trans)
        (do
          (println "  " (first t) "--[" (nth t 2) "->" (nth t 3) ":" (nth t 4) "]->" (second t))
          (print-transitions (rest trans))))))

(define (print-labels labels)
  (if (empty? labels)
      nil
      (let l (first labels)
        (do
          (println "  " (first l) ":" (second l))
          (print-labels (rest labels))))))

; ----------------------------------------------------------------------------
; CTL Model Checking (basic implementation)
; ----------------------------------------------------------------------------

; Check if a proposition holds in a state
(define (holds-at? k state prop)
  (member? prop (kripke-props-at k state)))

(define (member? item lst)
  (if (empty? lst)
      false
      (if (= item (first lst))
          true
          (member? item (rest lst)))))

; Evaluate a CTL formula at a state
; Returns true/false
(define (ctl-eval k state formula)
  (let formula-type (tag-type formula)
    (cond
      ((= formula-type 'ctl-prop)
       (holds-at? k state (tag-value formula)))
      
      ((= formula-type 'ctl-not)
       (not (ctl-eval k state (tag-value formula))))
      
      ((= formula-type 'ctl-and)
       (all? (lambda (f) (ctl-eval k state f)) (tag-value formula)))
      
      ((= formula-type 'ctl-or)
       (any? (lambda (f) (ctl-eval k state f)) (tag-value formula)))
      
      ((= formula-type 'ctl-implies)
       (let args (tag-value formula)
         (or (not (ctl-eval k state (first args)))
             (ctl-eval k state (second args)))))
      
      ((= formula-type 'ctl-EX)
       (ctl-EX-eval k state (tag-value formula)))
      
      ((= formula-type 'ctl-AX)
       (ctl-AX-eval k state (tag-value formula)))
      
      ((= formula-type 'ctl-EF)
       (ctl-EF-eval k state (tag-value formula) '()))
      
      ((= formula-type 'ctl-AF)
       (ctl-AF-eval k state (tag-value formula) '()))
      
      ((= formula-type 'ctl-EG)
       (ctl-EG-eval k state (tag-value formula) '()))
      
      ((= formula-type 'ctl-AG)
       (ctl-AG-eval k state (tag-value formula) '()))
      
      (true
       ; Unknown formula type or raw symbol (treat as prop)
       (if (symbol? formula)
           (holds-at? k state formula)
           false)))))

; EX φ - exists next state where φ holds
(define (ctl-EX-eval k state formula)
  (any? (lambda (s) (ctl-eval k s formula)) (kripke-successors k state)))

; AX φ - all next states satisfy φ  
(define (ctl-AX-eval k state formula)
  (let succs (kripke-successors k state)
    (if (empty? succs)
        true  ; vacuously true if no successors
        (all? (lambda (s) (ctl-eval k s formula)) succs))))

; EF φ - exists path where φ eventually holds (reachability)
(define (ctl-EF-eval k state formula visited)
  (if (member? state visited)
      false  ; cycle without finding φ
      (if (ctl-eval k state formula)
          true
          (any? (lambda (s) (ctl-EF-eval k s formula (cons state visited)))
                (kripke-successors k state)))))

; AF φ - on all paths, φ eventually holds
(define (ctl-AF-eval k state formula visited)
  (if (member? state visited)
      false  ; cycle without finding φ (counterexample)
      (if (ctl-eval k state formula)
          true
          (let succs (kripke-successors k state)
            (if (empty? succs)
                false  ; deadend without φ
                (all? (lambda (s) (ctl-AF-eval k s formula (cons state visited))) 
                      succs))))))

; EG φ - exists path where φ always holds
(define (ctl-EG-eval k state formula visited)
  (if (not (ctl-eval k state formula))
      false
      (if (member? state visited)
          true  ; found cycle where φ holds
          (let succs (kripke-successors k state)
            (if (empty? succs)
                true  ; φ holds at deadend
                (any? (lambda (s) (ctl-EG-eval k s formula (cons state visited)))
                      succs))))))

; AG φ - on all paths, φ always holds  
(define (ctl-AG-eval k state formula visited)
  (if (not (ctl-eval k state formula))
      false
      (if (member? state visited)
          true  ; cycle where φ holds
          (let succs (kripke-successors k state)
            (if (empty? succs)
                true
                (all? (lambda (s) (ctl-AG-eval k s formula (cons state visited)))
                      succs))))))

; ----------------------------------------------------------------------------
; High-level checking API
; ----------------------------------------------------------------------------

; Check a CTL property on a grammar
(define (check-property grammar-name property-name)
  (let k (grammar->kripke grammar-name)
    (if (nil? k)
        (do (println "Grammar not found:" grammar-name) nil)
        (let prop-data (get-property property-name)
          (if (nil? prop-data)
              (do (println "Property not found:" property-name) nil)
              (let formula (second prop-data)  ; skip property name
                (let initial (first (kripke-initial k))
                  (let result (ctl-eval k initial formula)
                    (do
                      (println "Checking" property-name "on" grammar-name)
                      (println "Formula:" formula)
                      (println "Result:" (if result "SATISFIED" "VIOLATED"))
                      result)))))))))

; Quick check a formula directly
(define (check-formula grammar-name formula)
  (let k (grammar->kripke grammar-name)
    (if (nil? k)
        (do (println "Grammar not found") nil)
        (let initial (first (kripke-initial k))
          (ctl-eval k initial formula)))))

; ----------------------------------------------------------------------------
(println "Kripke structure module loaded.")
