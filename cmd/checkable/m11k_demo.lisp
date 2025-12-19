; mm1k_demo_fixed.lisp
; Works with your interpreter semantics (global define, actor-only receive).

; ---------------- RNG ----------------

(define (rng_init! seed) (registry-set! "rng_seed" seed))

(define (rng_next_int!)
  (let* ((a 1103515245)
         (c 12345)
         (m 2147483648)
         (s (registry-get "rng_seed"))
         (s2 (mod (+ (* a s) c) m)))
    (begin
      (registry-set! "rng_seed" s2)
      s2)))

(define (rng_unit!)
  (/ (rng_next_int!) 2147483648))

(define (rng_flip p)
  (< (rng_unit!) p))

; ---------------- Queue EFSM (actor-local vars via set!) ----------------

(define (q_init cap pMu)
  (begin
    (set! q_x 0)
    (set! q_cap cap)
    (set! q_pMu pMu)
    (set! q_sumX 0)
    (set! q_ticks 0)
    (set! q_fullCount 0)
    (set! q_accepted 0)
    (set! q_departures 0)
    'ok))

(define (q_sample!)
  (begin
    (set! q_sumX (+ q_sumX q_x))
    (set! q_ticks (+ q_ticks 1))
    'ok))

(define (q_maybe_depart!)
  (if (and (> q_x 0) (rng_flip q_pMu))
      (begin
        (set! q_x (- q_x 1))
        (set! q_departures (+ q_departures 1))
        'ok)
      'ok))

(define (q_handle_arrival! t clientSym)
  (if (< q_x q_cap)
      (begin
        (set! q_x (+ q_x 1))
        (set! q_accepted (+ q_accepted 1))
        'ok)
      (begin
        (set! q_fullCount (+ q_fullCount 1))
        (send-to! clientSym (list 'full t))
        'ok)))

(define (q_drain_arrivals! t)
  (let msg (receive-now!)
    (if (and (symbol? msg) (= msg 'empty))
        'ok
        (begin
          (match msg
            ((list 'arrive ?client)
              (q_handle_arrival! t ?client))
            (_ 'ok))
          (q_drain_arrivals! t)))))

(define (q_step)
  (let msg (receive!)
    (match msg
      ((list 'tick ?t)
        (begin
          (q_sample!)
          (q_maybe_depart!)
          (q_drain_arrivals! ?t)
          (list 'continue (list 'q_step))))
      ((list 'stop)
        (begin
          (registry-set! "mm1k_sumX" q_sumX)
          (registry-set! "mm1k_ticks" q_ticks)
          (registry-set! "mm1k_avgX" (/ q_sumX q_ticks))
          (registry-set! "mm1k_fullCount" q_fullCount)
          (registry-set! "mm1k_accepted" q_accepted)
          (registry-set! "mm1k_departures" q_departures)
          (done!)))
      (_ (list 'continue (list 'q_step))))))

; ---------------- Client actor ----------------

(define (c_init clientSym queueSym pLambda)
  (begin
    (set! c_name clientSym)
    (set! c_queue queueSym)
    (set! c_pLambda pLambda)
    (set! c_sent 0)
    (set! c_rejected 0)
    'ok))

(define (c_maybe_arrive! t)
  (if (rng_flip c_pLambda)
      (begin
        (send-to! c_queue (list 'arrive c_name))
        (set! c_sent (+ c_sent 1))
        'ok)
      'ok))

(define (c_drain_rejects!)
  (let msg (receive-now!)
    (if (and (symbol? msg) (= msg 'empty))
        'ok
        (begin
          (match msg
            ((list 'full ?t)
              (set! c_rejected (+ c_rejected 1)))
            (_ 'ok))
          (c_drain_rejects!)))))

(define (c_step)
  (let msg (receive!)
    (match msg
      ((list 'tick ?t)
        (begin
          (c_maybe_arrive! ?t)
          (c_drain_rejects!)
          (list 'continue (list 'c_step))))
      ((list 'stop)
        (begin
          (registry-set! "mm1k_clientSent" c_sent)
          (registry-set! "mm1k_clientRejected" c_rejected)
          (done!)))
      (_ (list 'continue (list 'c_step))))))

; ---------------- Clock actor ----------------

(define (clk_init n queueSym clientSym)
  (begin
    (set! clk_n n)
    (set! clk_t 0)
    (set! clk_queue queueSym)
    (set! clk_client clientSym)
    'ok))

(define (clk_step)
  (if (>= clk_t clk_n)
      (begin
        (send-to! clk_queue (list 'stop))
        (send-to! clk_client (list 'stop))
        (done!))
      (begin
        (send-to! clk_queue (list 'tick clk_t))
        (send-to! clk_client (list 'tick clk_t))
        (set! clk_t (+ clk_t 1))
        (list 'continue (list 'clk_step)))))

; ---------------- Top-level run ----------------

(define (mm1k_run)
  (let* ((cap 10)
         (nTicks 5000)
         (pLambda 0.35)
         (pMu 0.40)
         (seed 1234567))
    (begin
      (rng_init! seed)

      (reset-scheduler)

      (spawn-actor 'queue  4096 (begin (q_init cap pMu) (list 'q_step)))
      (spawn-actor 'client 4096 (begin (c_init 'client 'queue pLambda) (list 'c_step)))
      (spawn-actor 'clock   64  (begin (clk_init nTicks 'queue 'client) (list 'clk_step)))

      (run-scheduler 500000)

      (println "=== M/M/1/K discrete sim ===")
      (println "avgX=" (registry-get "mm1k_avgX")
               " ticks=" (registry-get "mm1k_ticks"))
      (println "accepted=" (registry-get "mm1k_accepted")
               " rejected(full)=" (registry-get "mm1k_fullCount"))
      (println "departures=" (registry-get "mm1k_departures"))
      (println "clientSent=" (registry-get "mm1k_clientSent")
               " clientRejected=" (registry-get "mm1k_clientRejected"))

      (registry-get "mm1k_avgX"))))


(define (clk_step)
  (if (>= clk_t clk_n)
      (begin
        (send-to! clk_queue (list 'stop))
        (send-to! clk_client (list 'stop))
        (list 'continue (list 'clk_drain 10)))
      (begin
        (send-to! clk_queue (list 'tick clk_t))
        (send-to! clk_client (list 'tick clk_t))
        (set! clk_t (+ clk_t 1))
        (list 'continue (list 'clk_step)))))

(define (clk_drain k)
  (if (<= k 0)
      (done!)
      (begin
        (yield!)
        (list 'continue (list 'clk_drain (- k 1))))))


