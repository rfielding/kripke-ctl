; BoundedLISP test file

; Basic arithmetic
(println "=== Arithmetic ===")
(println "2 + 3 =" (+ 2 3))
(println "10 - 4 =" (- 10 4))
(println "6 * 7 =" (* 6 7))

; Lists and quotes
(println "\n=== Lists and Quotes ===")
(define mylist (' a b c d))
(println "mylist:" mylist)
(println "first:" (first mylist))
(println "rest:" (rest mylist))

; Functions
(println "\n=== Functions ===")
(define (square x) (* x x))
(println "square 5 =" (square 5))

; Tail recursion
(println "\n=== Tail Recursion ===")
(define (sum-to n acc)
  (if (= n 0)
      acc
      (tail sum-to (- n 1) (+ acc n))))
(println "sum 1..100 =" (sum-to 100 0))

; Bounded stack
(println "\n=== Bounded Stack ===")
(define s (make-stack 4))
(println "created stack capacity 4")
(push-now! s 10)
(push-now! s 20)
(push-now! s 30)
(println "pushed 10, 20, 30")
(println "stack-read 0:" (stack-read s 0))
(println "stack-read 1:" (stack-read s 1))
(println "stack-read 2:" (stack-read s 2))
(println "pop:" (pop-now! s))
(println "pop:" (pop-now! s))

; Bounded queue
(println "\n=== Bounded Queue ===")
(define q (make-queue 3))
(send-now! q "first")
(send-now! q "second")
(send-now! q "third")
(println "queue full?" (queue-full? q))
(println "recv:" (recv-now! q))
(println "recv:" (recv-now! q))
(println "recv:" (recv-now! q))
(println "recv empty:" (recv-now! q))

; Pattern matching
(println "\n=== Pattern Matching ===")
(define (describe x)
  (match x
    (0 "zero")
    (1 "one")
    ((_ _ _) "three element list")
    (_ "something else")))

(println "describe 0:" (describe 0))
(println "describe 1:" (describe 1))
(println "describe '(a b c):" (describe (' a b c)))
(println "describe 42:" (describe 42))

; Tokenizer-like tail recursion
(println "\n=== Tokenizer Pattern ===")
(define (count-down n)
  (if (<= n 0)
      (println "done!")
      (do
        (println "count:" n)
        (tail count-down (- n 1)))))
(count-down 5)

(println "\n=== All tests complete ===")
