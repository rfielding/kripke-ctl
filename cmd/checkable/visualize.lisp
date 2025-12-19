; ============================================================================
; Visualization Module
; Generates Mermaid diagrams from grammars and Kripke structures
; ============================================================================

; ----------------------------------------------------------------------------
; State Diagram Generation
; ----------------------------------------------------------------------------

; Generate Mermaid state diagram from Kripke structure
(define (kripke->mermaid k)
  (if (not (kripke? k))
      "Error: not a Kripke structure"
      (string-append
        "stateDiagram-v2\n"
        (string-append "    [*] --> " (symbol->string (first (kripke-initial k))) "\n")
        (mermaid-trans-str (kripke-transitions k))
        (mermaid-term-str (kripke-states k) (kripke-transitions k)))))

(define (mermaid-trans-str trans)
  (if (empty? trans)
      ""
      (let t (first trans)
        (string-append
          "    " (symbol->string (first t)) 
          " --> " (symbol->string (second t))
          " : " (symbol->string (nth t 4)) "\n"
          (mermaid-trans-str (rest trans))))))

(define (mermaid-term-str states transitions)
  (if (empty? states)
      ""
      (let s (first states)
        (let outgoing (filter (lambda (t) (= (first t) s)) transitions)
          (string-append
            (if (empty? outgoing)
                (string-append "    " (symbol->string s) " --> [*]\n")
                "")
            (mermaid-term-str (rest states) transitions))))))

; Convenience: grammar name -> state diagram
(define (grammar->state-diagram grammar-name)
  (let k (grammar->kripke grammar-name)
    (if (nil? k)
        "Error: grammar not found"
        (kripke->mermaid k))))

; ----------------------------------------------------------------------------
; Sequence Diagram Generation  
; ----------------------------------------------------------------------------

; Generate sequence diagram showing all message types
(define (grammar->sequence grammar-name)
  (let g (get-grammar grammar-name)
    (if (nil? g)
        "Error: grammar not found"
        (let trans (grammar-transitions g)
          (let roles (unique (collect-roles trans))
            (string-append
              "sequenceDiagram\n"
              (sequence-participants roles)
              (sequence-trans-str trans)))))))

(define (collect-roles trans)
  (if (empty? trans)
      '()
      (let t (first trans)
        (cons (nth t 2) (cons (nth t 3) (collect-roles (rest trans)))))))

(define (unique lst)
  (unique-helper lst '()))

(define (unique-helper lst acc)
  (if (empty? lst)
      acc
      (let item (first lst)
        (if (member? item acc)
            (unique-helper (rest lst) acc)
            (unique-helper (rest lst) (append acc (list item)))))))

(define (sequence-participants roles)
  (if (empty? roles)
      ""
      (string-append
        "    participant " (symbol->string (first roles)) "\n"
        (sequence-participants (rest roles)))))

(define (sequence-trans-str trans)
  (if (empty? trans)
      ""
      (let t (first trans)
        (string-append
          "    " (symbol->string (nth t 2)) 
          " ->> " (symbol->string (nth t 3))
          " : " (symbol->string (nth t 4)) "\n"
          (sequence-trans-str (rest trans))))))

; ----------------------------------------------------------------------------
; Flowchart Generation
; ----------------------------------------------------------------------------

; Generate flowchart showing protocol flow
(define (grammar->flowchart grammar-name)
  (let g (get-grammar grammar-name)
    (if (nil? g)
        "Error: grammar not found"
        (let trans (grammar-transitions g)
          (let states (grammar-states g)
            (string-append
              "flowchart TD\n"
              (flowchart-nodes states trans)
              (flowchart-edges trans)))))))

(define (flowchart-nodes states trans)
  (if (empty? states)
      ""
      (let s (first states)
        (let sname (symbol->string s)
          (let outgoing (filter (lambda (t) (= (first t) s)) trans)
            (let incoming (filter (lambda (t) (= (second t) s)) trans)
              (let shape (cond
                           ((empty? outgoing) (string-append "([" sname "])"))
                           ((empty? incoming) (string-append "((" sname "))"))
                           (true (string-append "[" sname "]")))
                (string-append "    " sname shape "\n"
                              (flowchart-nodes (rest states) trans))))))))))

(define (flowchart-edges trans)
  (if (empty? trans)
      ""
      (let t (first trans)
        (string-append
          "    " (symbol->string (first t))
          " -->|" (symbol->string (nth t 4)) "| "
          (symbol->string (second t)) "\n"
          (flowchart-edges (rest trans))))))

; ----------------------------------------------------------------------------
; Pie Chart (for simple data)
; ----------------------------------------------------------------------------

(define (make-pie-chart title data)
  ; data is list of (label value) pairs
  (string-append
    "pie title " title "\n"
    (pie-data-str data)))

(define (pie-data-str data)
  (if (empty? data)
      ""
      (let item (first data)
        (string-append
          "    \"" (symbol->string (first item)) "\" : " 
          (number->string (second item)) "\n"
          (pie-data-str (rest data))))))

; ----------------------------------------------------------------------------
; Bar Chart (using XY chart)
; ----------------------------------------------------------------------------

(define (make-bar-chart title x-label y-label data)
  ; data is list of (label value) pairs
  (string-append
    "xychart-beta\n"
    "    title \"" title "\"\n"
    "    x-axis " (bar-labels data) "\n"
    "    y-axis \"" y-label "\"\n"
    "    bar " (bar-values data) "\n"))

(define (bar-labels data)
  (string-append "[" (bar-labels-inner data) "]"))

(define (bar-labels-inner data)
  (if (empty? data)
      ""
      (if (empty? (rest data))
          (string-append "\"" (symbol->string (first (first data))) "\"")
          (string-append "\"" (symbol->string (first (first data))) "\", " 
                        (bar-labels-inner (rest data))))))

(define (bar-values data)
  (string-append "[" (bar-values-inner data) "]"))

(define (bar-values-inner data)
  (if (empty? data)
      ""
      (if (empty? (rest data))
          (number->string (second (first data)))
          (string-append (number->string (second (first data))) ", "
                        (bar-values-inner (rest data))))))

; ----------------------------------------------------------------------------
(println "Visualization module loaded.")
