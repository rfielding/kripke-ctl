; ============================================================================
; BoundedLISP Prologue
; High-level constructs for formal specification
; ============================================================================

; ----------------------------------------------------------------------------
; Grammar Definition
; ----------------------------------------------------------------------------

; (defgrammar name
;   (State1
;     (Role1 -> Role2 : message) -> State2
;     (Role1 -> Role2 : other)   -> State3)
;   (State2 ...)
;   ...)

(define (defgrammar name . body)
  (let data (tag 'grammar (cons name body))
    (do
      (registry-set! name data)
      data)))

(define (grammar? name)
  (let g (registry-get name)
    (if (nil? g)
        false
        (tag-is? g 'grammar))))

(define (get-grammar name)
  (let g (registry-get name)
    (if (tag-is? g 'grammar)
        (tag-value g)
        nil)))

(define (grammar-name g)
  (first (tag-value g)))

(define (grammar-states g)
  (rest (tag-value g)))

; ----------------------------------------------------------------------------
; Actor Definition  
; ----------------------------------------------------------------------------

; (defactor name
;   (follows grammar-name :as Role)
;   (mailbox-size 16)
;   (state ...))

(define (defactor name . body)
  (let mailbox (make-queue 16)
    (let data (tag 'actor (list name mailbox body))
      (do
        (registry-set! name data)
        data))))

(define (actor? name)
  (let a (registry-get name)
    (if (nil? a)
        false
        (tag-is? a 'actor))))

(define (get-actor name)
  (let a (registry-get name)
    (if (tag-is? a 'actor)
        (tag-value a)
        nil)))

(define (actor-name a)
  (first (tag-value a)))

(define (actor-mailbox a)
  (nth (tag-value a) 1))

(define (actor-body a)
  (nth (tag-value a) 2))

; ----------------------------------------------------------------------------
; Property Definition (CTL formulas)
; ----------------------------------------------------------------------------

; (defproperty name
;   (AG (-> filed (AF approved))))

(define (defproperty name . body)
  (let data (tag 'property (cons name body))
    (do
      (registry-set! name data)
      data)))

(define (property? name)
  (let p (registry-get name)
    (if (nil? p)
        false
        (tag-is? p 'property))))

(define (get-property name)
  (let p (registry-get name)
    (if (tag-is? p 'property)
        (tag-value p)
        nil)))

; ----------------------------------------------------------------------------
; CTL Formula Constructors
; ----------------------------------------------------------------------------

; Atomic proposition
(define (prop name) 
  (tag 'ctl-prop name))

; Boolean connectives
(define (ctl-and . args)
  (tag 'ctl-and args))

(define (ctl-or . args)
  (tag 'ctl-or args))

(define (ctl-not formula)
  (tag 'ctl-not formula))

(define (ctl-implies p q)
  (tag 'ctl-implies (list p q)))

; Universal path quantifiers
(define (AX formula)
  (tag 'ctl-AX formula))

(define (AF formula)
  (tag 'ctl-AF formula))

(define (AG formula)
  (tag 'ctl-AG formula))

(define (AU p q)
  (tag 'ctl-AU (list p q)))

; Existential path quantifiers
(define (EX formula)
  (tag 'ctl-EX formula))

(define (EF formula)
  (tag 'ctl-EF formula))

(define (EG formula)
  (tag 'ctl-EG formula))

(define (EU p q)
  (tag 'ctl-EU (list p q)))

; ----------------------------------------------------------------------------
; Role Projection Helpers
; ----------------------------------------------------------------------------

; Extract all roles mentioned in a grammar
(define (grammar-roles grammar)
  ; Walk the grammar and collect unique role names
  ; For now, placeholder - returns empty list
  '())

; Project a grammar to a specific role's view
(define (project-to-role grammar role)
  ; Filter transitions to only those involving this role
  ; Returns a local FSM
  ; For now, placeholder
  nil)

; ----------------------------------------------------------------------------
; Instance Management  
; ----------------------------------------------------------------------------

; Create a running instance of a protocol
(define (spawn-protocol grammar-name bindings)
  (let instance-id (gensym 'instance)
    (let data (tag 'instance (list instance-id grammar-name bindings 'Start))
      (do
        (registry-set! instance-id data)
        instance-id))))

; Get current state of an instance
(define (instance-state instance-id)
  (let inst (registry-get instance-id)
    (if (tag-is? inst 'instance)
        (nth (tag-value inst) 3)
        nil)))

; ----------------------------------------------------------------------------
; Utilities
; ----------------------------------------------------------------------------

; List all registered items by type
(define (list-grammars)
  (filter-registry 'grammar))

(define (list-actors)
  (filter-registry 'actor))

(define (list-properties)
  (filter-registry 'property))

(define (filter-registry type-tag)
  (filter-by-tag (registry-keys) type-tag))

(define (filter-by-tag keys type-tag)
  (if (empty? keys)
      '()
      (let k (first keys)
        (let v (registry-get k)
          (if (tag-is? v type-tag)
              (cons k (filter-by-tag (rest keys) type-tag))
              (filter-by-tag (rest keys) type-tag))))))

; Pretty print a grammar
(define (print-grammar name)
  (let g (get-grammar name)
    (if (nil? g)
        (println "Grammar not found:" name)
        (do
          (println "Grammar:" (first g))
          (print-states (rest g))))))

(define (print-states states)
  (if (empty? states)
      nil
      (do
        (println "  State:" (first states))
        (print-states (rest states)))))

; ----------------------------------------------------------------------------
; End of Prologue
; ----------------------------------------------------------------------------

(println "BoundedLISP Prologue loaded.")
