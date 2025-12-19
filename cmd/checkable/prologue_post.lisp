; =========================
; Minimal unit test harness
; =========================

(define *tests* (list))
(define *test-failures* (list))
(define *test-runs* 0)

(define (test name desc thunk)
  ; store (name desc thunk)
  (set! *tests* (cons (list name desc thunk) *tests*))
  'ok)

(define (fail name desc got expected)
  (set! *test-failures* (cons (list name desc got expected) *test-failures*))
  'fail)

(define (assert-eq name desc got expected)
  (set! *test-runs* (+ *test-runs* 1))
  (if (= got expected)
      'ok
      (fail name desc got expected)))

(define (assert-true name desc v)
  (set! *test-runs* (+ *test-runs* 1))
  (if v
      'ok
      (fail name desc v true)))

(define (assert-false name desc v)
  (set! *test-runs* (+ *test-runs* 1))
  (if (not v)
      'ok
      (fail name desc v false)))

(define (run-tests!)
  ; reverse so tests run in declared order
  (set! *test-failures* (list))
  (let tests (reverse *tests*)
    (begin
      (println "=== Running prologue tests ===")
      (do-each tests
        (fn (t)
          (let name (nth t 0)
            (let desc (nth t 1)
              (let thunk (nth t 2)
                (thunk))))))
      (if (empty? *test-failures*)
          (begin
            (println "PASS:" *test-runs* "assertions")
            true)
          (begin
            (println "FAIL:" (length *test-failures*) "failures out of" *test-runs* "assertions")
            (do-each (reverse *test-failures*)
              (fn (f)
                (println " - " (nth f 0) ": " (nth f 1)
                         " got=" (repr (nth f 2))
                         " expected=" (repr (nth f 3))))))
            false)))))

; helper: reverse + do-each if you don’t already have them in prologue
(define (reverse xs)
  (define (go xs acc)
    (if (empty? xs)
        acc
        (tail go (rest xs) (cons (first xs) acc))))
  (go xs (list)))

(define (do-each xs f)
  (define (go xs)
    (if (empty? xs)
        'ok
        (begin (f (first xs))
               (tail go (rest xs)))))
  (go xs))

