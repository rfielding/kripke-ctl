; ============================================================================
; Probability Distribution Transformations
; Transform uniform [0,1] pre-rolls into various distributions
; ============================================================================

; ----------------------------------------------------------------------------
; Core Distributions
; ----------------------------------------------------------------------------

; Exponential distribution: models inter-arrival times, service times
; Given uniform U in (0,1], returns exponential with rate λ
; Mean = 1/rate
(define (exponential u rate)
  (/ (- 0 (ln u)) rate))

; Alias for clarity
(define (exponential-from-uniform u rate)
  (exponential u rate))

; Poisson process: given a rate λ and uniform u, get next arrival time
; This is equivalent to exponential(λ)
(define (poisson-arrival u rate)
  (exponential u rate))

; Geometric distribution: number of failures before first success
; p = probability of success on each trial
; Returns count (0, 1, 2, ...)
(define (geometric u p)
  (floor (/ (ln u) (ln (- 1 p)))))

; Bernoulli trial: success (1) or failure (0) with probability p
(define (bernoulli u p)
  (if (< u p) 1 0))

; Uniform in range [a, b]
(define (uniform-range u a b)
  (+ a (* u (- b a))))

; ----------------------------------------------------------------------------
; Normal/Gaussian Distribution (Box-Muller transform)
; Requires TWO uniform values
; ----------------------------------------------------------------------------

; Standard normal (mean=0, stddev=1)
; Uses Box-Muller transform: given U1, U2 uniform
; Z0 = sqrt(-2*ln(U1)) * cos(2*π*U2)
; Z1 = sqrt(-2*ln(U1)) * sin(2*π*U2)
(define *pi* 3.141592653589793)

(define (normal-z0 u1 u2)
  (* (sqrt (* -2 (ln u1))) 
     (cos (* 2 *pi* u2))))

(define (normal-z1 u1 u2)
  (* (sqrt (* -2 (ln u1))) 
     (sin (* 2 *pi* u2))))

; General normal with mean μ and stddev σ
(define (normal u1 u2 mean stddev)
  (+ mean (* stddev (normal-z0 u1 u2))))

; ----------------------------------------------------------------------------
; Discrete Distributions
; ----------------------------------------------------------------------------

; Discrete uniform: pick integer in [min, max] inclusive
(define (discrete-uniform u min-val max-val)
  (+ min-val (floor (* u (+ 1 (- max-val min-val))))))

; Categorical/multinomial: pick from weighted options
; weights is a list of (weight . value) pairs
; Returns the value whose cumulative weight bracket contains u
(define (categorical u weights)
  (let total (sum-weights weights)
    (categorical-pick u weights 0 total)))

(define (sum-weights weights)
  (if (empty? weights)
      0
      (+ (first (first weights)) (sum-weights (rest weights)))))

(define (categorical-pick u weights cumulative total)
  (if (empty? weights)
      nil
      (let w (first weights)
        (let new-cumulative (+ cumulative (/ (first w) total))
          (if (< u new-cumulative)
              (second w)
              (categorical-pick u (rest weights) new-cumulative total))))))

; Weighted coin flip helper
(define (weighted-choice u p val-true val-false)
  (if (< u p) val-true val-false))

; ----------------------------------------------------------------------------
; Queueing Theory Helpers
; ----------------------------------------------------------------------------

; M/M/1 queue parameters
; λ = arrival rate (arrivals per time unit)
; μ = service rate (services per time unit)  
; ρ = λ/μ = utilization (must be < 1 for stability)

; Generate next inter-arrival time
(define (mm1-arrival-time u lambda)
  (exponential u lambda))

; Generate next service time  
(define (mm1-service-time u mu)
  (exponential u mu))

; Theoretical M/M/1 metrics for comparison
(define (mm1-utilization lambda mu)
  (/ lambda mu))

(define (mm1-avg-queue-length lambda mu)
  (let rho (/ lambda mu)
    (/ (* rho rho) (- 1 rho))))

(define (mm1-avg-system-length lambda mu)
  (let rho (/ lambda mu)
    (/ rho (- 1 rho))))

(define (mm1-avg-wait-time lambda mu)
  (let rho (/ lambda mu)
    (/ rho (* mu (- 1 rho)))))

(define (mm1-avg-system-time lambda mu)
  (/ 1 (- mu lambda)))

; ----------------------------------------------------------------------------
; Dice Pool Management
; ----------------------------------------------------------------------------

; A dice pool is a bounded queue of pre-rolled uniform values
; Create with: (make-dice-pool size rolls)

(define (make-dice-pool size initial-rolls)
  (let pool (make-queue size)
    (do
      (fill-dice-pool pool initial-rolls)
      pool)))

(define (fill-dice-pool pool rolls)
  (if (empty? rolls)
      pool
      (do
        (enqueue-now! pool (first rolls))
        (fill-dice-pool pool (rest rolls)))))

; Get next die from pool (non-blocking, returns 'empty if exhausted)
(define (next-die! pool)
  (dequeue-now! pool))

; Check if pool has dice remaining
(define (dice-remaining? pool)
  (not (queue-empty? pool)))

; ----------------------------------------------------------------------------
(println "Distributions module loaded.")
