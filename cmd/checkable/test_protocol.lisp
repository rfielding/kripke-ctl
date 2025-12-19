; Test multi-party protocol: Acquisition
; Acquirer -> Target -> Regulator choreography

(println "=== Multi-Party Protocol Test ===\n")

; ============================================================================
; Protocol states as behaviors (simplified for debugging)
; ============================================================================

; Acquirer: send offer, wait for accept/reject
(define (acquirer-offer)
  (do
    (println "Acquirer: sending offer to Target")
    (send-to! 'target '(offer 1000000))
    (list 'become '(acquirer-wait-response))))

(define (acquirer-wait-response)
  (let msg (receive-now!)
    (if (= msg 'empty)
        'yield
        (do
          (println "Acquirer: got response:" msg)
          (if (= (first msg) 'accept)
              (do
                (println "Acquirer: offer accepted, filing with regulator")
                (send-to! 'regulator '(filing details))
                (list 'become '(acquirer-wait-approval)))
              (do
                (println "Acquirer: offer rejected")
                'done))))))

(define (acquirer-wait-approval)
  (let msg (receive-now!)
    (if (= msg 'empty)
        'yield
        (do
          (println "Acquirer: got from regulator:" msg)
          (if (= (first msg) 'approved)
              (do
                (println "Acquirer: approved! sending payment")
                (send-to! 'target '(payment 1000000))
                (list 'become '(acquirer-wait-assets)))
              (do
                (println "Acquirer: denied")
                'done))))))

(define (acquirer-wait-assets)
  (let msg (receive-now!)
    (if (= msg 'empty)
        'yield
        (do
          (println "Acquirer: got:" msg)
          (println "Acquirer: acquisition complete!")
          'done))))

; Target: wait for offer, then payment
(define (target-wait-offer)
  (let msg (receive-now!)
    (if (= msg 'empty)
        'yield
        (do
          (println "Target: got offer:" msg)
          (println "Target: accepting")
          (send-to! 'acquirer '(accept))
          (list 'become '(target-wait-payment))))))

(define (target-wait-payment)
  (let msg (receive-now!)
    (if (= msg 'empty)
        'yield
        (do
          (println "Target: got payment:" msg)
          (println "Target: transferring assets")
          (send-to! 'acquirer '(assets))
          'done))))

; Regulator: wait for filing, approve
(define (regulator-wait-filing)
  (let msg (receive-now!)
    (if (= msg 'empty)
        'yield
        (do
          (println "Regulator: got filing:" msg)
          (println "Regulator: approving")
          (send-to! 'acquirer '(approved))
          'done))))

; ============================================================================
; Spawn actors and run
; ============================================================================

(println "Spawning actors...")
(spawn-actor 'acquirer 8 '(acquirer-offer))
(spawn-actor 'target 8 '(target-wait-offer))
(spawn-actor 'regulator 8 '(regulator-wait-filing))

(println "Actors:" (list-actors-sched))
(set-trace! false)  ; Set to true for detailed trace

(println "\nRunning acquisition protocol...")
(define result (run-scheduler 50))
(println "Result:" result)
