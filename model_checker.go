package main

// ModelChecker implements the CTL model checking algorithm
type ModelChecker struct {
	Structure *KripkeStructure
}

// NewModelChecker creates a new model checker
func NewModelChecker(k *KripkeStructure) *ModelChecker {
	return &ModelChecker{Structure: k}
}

// Check evaluates a CTL formula on the Kripke structure and returns the set of states where it holds
func (mc *ModelChecker) Check(formula CTLFormula) map[State]bool {
	switch f := formula.(type) {
	case True:
		return mc.checkTrue()
	case False:
		return mc.checkFalse()
	case AtomicProp:
		return mc.checkAtomicProp(f)
	case Not:
		return mc.checkNot(f)
	case And:
		return mc.checkAnd(f)
	case Or:
		return mc.checkOr(f)
	case Implies:
		return mc.checkImplies(f)
	case EX:
		return mc.checkEX(f)
	case AX:
		return mc.checkAX(f)
	case EF:
		return mc.checkEF(f)
	case AF:
		return mc.checkAF(f)
	case EG:
		return mc.checkEG(f)
	case AG:
		return mc.checkAG(f)
	case EU:
		return mc.checkEU(f)
	case AU:
		return mc.checkAU(f)
	default:
		return make(map[State]bool)
	}
}

// Holds checks if a formula holds in the initial state
func (mc *ModelChecker) Holds(formula CTLFormula) bool {
	satisfyingStates := mc.Check(formula)
	return satisfyingStates[mc.Structure.InitialState]
}

func (mc *ModelChecker) checkTrue() map[State]bool {
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		result[state] = true
	}
	return result
}

func (mc *ModelChecker) checkFalse() map[State]bool {
	return make(map[State]bool)
}

func (mc *ModelChecker) checkAtomicProp(f AtomicProp) map[State]bool {
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		if mc.Structure.HasLabel(state, f.Prop) {
			result[state] = true
		}
	}
	return result
}

func (mc *ModelChecker) checkNot(f Not) map[State]bool {
	subResult := mc.Check(f.Formula)
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		if !subResult[state] {
			result[state] = true
		}
	}
	return result
}

func (mc *ModelChecker) checkAnd(f And) map[State]bool {
	leftResult := mc.Check(f.Left)
	rightResult := mc.Check(f.Right)
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		if leftResult[state] && rightResult[state] {
			result[state] = true
		}
	}
	return result
}

func (mc *ModelChecker) checkOr(f Or) map[State]bool {
	leftResult := mc.Check(f.Left)
	rightResult := mc.Check(f.Right)
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		if leftResult[state] || rightResult[state] {
			result[state] = true
		}
	}
	return result
}

func (mc *ModelChecker) checkImplies(f Implies) map[State]bool {
	// p -> q is equivalent to ¬p ∨ q
	return mc.Check(Or{Not{f.Left}, f.Right})
}

func (mc *ModelChecker) checkEX(f EX) map[State]bool {
	subResult := mc.Check(f.Formula)
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		successors := mc.Structure.GetSuccessors(state)
		for _, succ := range successors {
			if subResult[succ] {
				result[state] = true
				break
			}
		}
	}
	return result
}

func (mc *ModelChecker) checkAX(f AX) map[State]bool {
	subResult := mc.Check(f.Formula)
	result := make(map[State]bool)
	for _, state := range mc.Structure.States {
		successors := mc.Structure.GetSuccessors(state)
		if len(successors) == 0 {
			continue
		}
		allSatisfy := true
		for _, succ := range successors {
			if !subResult[succ] {
				allSatisfy = false
				break
			}
		}
		if allSatisfy {
			result[state] = true
		}
	}
	return result
}

func (mc *ModelChecker) checkEF(f EF) map[State]bool {
	// EF p = E[true U p]
	return mc.checkEU(EU{True{}, f.Formula})
}

func (mc *ModelChecker) checkAF(f AF) map[State]bool {
	// AF p = ¬EG ¬p
	return mc.Check(Not{EG{Not{f.Formula}}})
}

func (mc *ModelChecker) checkEG(f EG) map[State]bool {
	// Fixed point algorithm: start with all states satisfying the subformula
	subResult := mc.Check(f.Formula)
	result := make(map[State]bool)
	for state := range subResult {
		result[state] = true
	}
	
	changed := true
	for changed {
		changed = false
		newResult := make(map[State]bool)
		for state := range result {
			// State stays in result if it has at least one successor in result
			successors := mc.Structure.GetSuccessors(state)
			hasSuccessorInResult := false
			for _, succ := range successors {
				if result[succ] {
					hasSuccessorInResult = true
					break
				}
			}
			if hasSuccessorInResult {
				newResult[state] = true
			} else {
				changed = true
			}
		}
		result = newResult
	}
	return result
}

func (mc *ModelChecker) checkAG(f AG) map[State]bool {
	// AG p = ¬EF ¬p
	return mc.Check(Not{EF{Not{f.Formula}}})
}

func (mc *ModelChecker) checkEU(f EU) map[State]bool {
	// Fixed point algorithm
	leftResult := mc.Check(f.Left)
	rightResult := mc.Check(f.Right)
	
	result := make(map[State]bool)
	for state := range rightResult {
		result[state] = true
	}
	
	changed := true
	for changed {
		changed = false
		for _, state := range mc.Structure.States {
			if result[state] {
				continue
			}
			if !leftResult[state] {
				continue
			}
			// Check if there's a successor in result
			successors := mc.Structure.GetSuccessors(state)
			for _, succ := range successors {
				if result[succ] {
					result[state] = true
					changed = true
					break
				}
			}
		}
	}
	return result
}

func (mc *ModelChecker) checkAU(f AU) map[State]bool {
	// Fixed point algorithm
	leftResult := mc.Check(f.Left)
	rightResult := mc.Check(f.Right)
	
	result := make(map[State]bool)
	for state := range rightResult {
		result[state] = true
	}
	
	changed := true
	for changed {
		changed = false
		for _, state := range mc.Structure.States {
			if result[state] {
				continue
			}
			if !leftResult[state] {
				continue
			}
			// Check if all successors are in result
			successors := mc.Structure.GetSuccessors(state)
			if len(successors) == 0 {
				continue
			}
			allInResult := true
			for _, succ := range successors {
				if !result[succ] {
					allInResult = false
					break
				}
			}
			if allInResult {
				result[state] = true
				changed = true
			}
		}
	}
	return result
}
