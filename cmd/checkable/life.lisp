(system
  :name "HumanLifetime"
  :english "Models one human life from birth to death with bounded time (ticks=years). At 20 enters workforce, at 60 retires, at 80 dies. Before 20, must either complete school or learn a trade. Resurrection is modeled as impossible (no Dead->Alive transition)."

  (time
    :tick-unit "year"
    :max-tick 80)

  ;; --------------------------
  ;; Propositions (with English)
  ;; --------------------------
  (prop Alive
    :english "Human is alive."
    :holds (== human.status 'alive))

  (prop Dead
    :english "Human is dead."
    :holds (== human.status 'dead))

  (prop InWorkforce
    :english "Human is in the workforce (age >=20 and <60)."
    :holds (and (>= human.age 20) (< human.age 60) (== human.phase 'working)))

  (prop Retired
    :english "Human is retired (age >=60 and <80)."
    :holds (and (>= human.age 60) (< human.age 80) (== human.phase 'retired)))

  (prop HasSkill
    :english "Human has fulfilled family obligation by acquiring either education or a trade."
    :holds (== human.hasSkill true))

  (prop Resurrected
    :english "Human becomes alive again after having been dead."
    :holds (and (== human.status 'alive) (== human.wasDead true)))

  (prop PathSchool
    :english "Artifact: chosen path is school."
    :holds (== human.path 'school))

  (prop PathTrade
    :english "Artifact: chosen path is trade."
    :holds (== human.path 'trade))

  (prop Youth
    :english "Artifact: life-phase partition state youth (<20 while alive)."
    :holds (and (== human.status 'alive) (< human.age 20)))

  (prop WorkingAge
    :english "Artifact: life-phase partition state working (20..59 while alive)."
    :holds (and (== human.status 'alive) (>= human.age 20) (< human.age 60)))

  (prop RetiredAge
    :english "Artifact: life-phase partition state retired (60..79 while alive)."
    :holds (and (== human.status 'alive) (>= human.age 60) (< human.age 80)))

  ;; --------------------------
  ;; Actor: Human (single EFSM)
  ;; --------------------------
  (actor human
    :english "Single-threaded EFSM with chance node for choosing education vs trade, deterministic aging, and hard death at 80."
    (vars
      (age 0)
      (status 'alive)      ;; 'alive | 'dead
      (phase 'child)       ;; 'child | 'working | 'retired
      (hasSkill false)
      (wasDead false)      ;; once dead, stays true
      (path 'unset))       ;; 'school | 'trade

    (efsm
      ;; Guard partition states (nested predicates)
      (pred-state life
        :english "Top-level partition by alive/dead."
        (case
          ((== status 'alive) alive)
          (else dead)))

      (pred-state alive
        :english "Partition of alive life by age milestones."
        (case
          ((< age 20) youth)
          ((< age 60) working)
          ((< age 80) retired)
          (else dying)))

      (state youth
        :english "Before 20: must acquire education or trade."
        (steps
          ;; choose path once, around age 16, nondeterministically (chance node)
          (chance choose_path
            :when (and (== path 'unset) (>= age 16))
            (outcomes
              (0.5 (code (set path 'school)))
              (0.5 (code (set path 'trade)))))

          ;; progress toward skill (deterministic once path chosen)
          (code learn
            :when (and (== hasSkill false) (== path 'school) (>= age 18))
            (set hasSkill true))

          (code apprentice
            :when (and (== hasSkill false) (== path 'trade) (>= age 18))
            (set hasSkill true))

          ;; age advances each year while alive
          (code age_step
            :when true
            (set age (+ age 1)))

          ;; update phase at 20/60
          (code phase_step
            :when true
            (if (>= age 20) (set phase 'working) (set phase 'child)))))

      (state working
        :english "20..59: workforce."
        (steps
          (code age_step :when true (set age (+ age 1)))
          (code phase_step :when true
            (if (>= age 60) (set phase 'retired) (set phase 'working)))))

      (state retired
        :english "60..79: retired."
        (steps
          (code age_step :when true (set age (+ age 1)))
          (code phase_step :when true
            (if (>= age 80) (set phase 'deadSoon) (set phase 'retired)))))

      (state dying
        :english "At age >=80: death occurs."
        (steps
          (code die
            :when true
            (set status 'dead)
            (set wasDead true)))))

      (state dead
        :english "Dead is absorbing; resurrection is impossible in this model."
        (steps
          ;; no transitions back to alive
          (code stay_dead :when true (noop))))))

  ;; --------------------------
  ;; Requirements / Questions
  ;; --------------------------
  (require Q1
    :ctl (AG (-> (and Alive (< human.age 20)) (AF HasSkill)))
    :english "While alive and before 20, it is required that eventually the human acquires a skill (school or trade).")

  (require Q2
    :ctl (AG (-> (== human.age 20) HasSkill))
    :english "At age 20, the human must already have a skill (stronger check).")

  (require Q3
    :ctl (AG (-> (== human.age 80) Dead))
    :english "At age 80, the human is dead (in this model).")

  (require Q4
    :ctl (EF Resurrected)
    :english "Is it possible to resurrect him? (Should be false here because dead is absorbing.)")

  (require A1
    :ctl (EF PathSchool)
    :english "Artifact: there exists a run where the human chooses school.")

  (require A2
    :ctl (EF PathTrade)
    :english "Artifact: there exists a run where the human chooses trade.")

  (require A3
    :ctl (AG (-> PathSchool (AG (-> (== human.path 'school) (AX (== human.path 'school))))))
    :english "Artifact: once path is school, it remains school forever (no switching).")

  (require A4
    :ctl (AG (-> PathTrade (AG (-> (== human.path 'trade) (AX (== human.path 'trade))))))
    :english "Artifact: once path is trade, it remains trade forever (no switching).")

  (require A5
    :ctl (AG (-> Alive (or Youth (or WorkingAge RetiredAge))))
    :english "Artifact: while alive, age-partition is always youth or working or retired (never outside these until dying/dead)."))
