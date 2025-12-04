package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== Kripke-CTL: CTL Model Checker with OpenAI Integration ===")
	fmt.Println()

	// Check if OpenAI API key is available
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""
	if hasOpenAI {
		fmt.Println("✓ OpenAI API key detected")
	} else {
		fmt.Println("⚠ OpenAI API key not found (set OPENAI_API_KEY environment variable for AI features)")
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Options:")
		fmt.Println("1. Use predefined example (Traffic Light)")
		fmt.Println("2. Use predefined example (Mutual Exclusion)")
		fmt.Println("3. Use predefined example (Simple)")
		if hasOpenAI {
			fmt.Println("4. Generate model from English description (OpenAI)")
		}
		fmt.Println("5. Exit")
		fmt.Print("\nSelect option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var k *KripkeStructure

		switch input {
		case "1":
			k = CreateTrafficLightExample()
			fmt.Println("\n=== Traffic Light Model ===")
			fmt.Println(k)
			runModelChecking(k, reader)

		case "2":
			k = CreateMutualExclusionExample()
			fmt.Println("\n=== Mutual Exclusion Model ===")
			fmt.Println(k)
			runModelChecking(k, reader)

		case "3":
			k = CreateSimpleExample()
			fmt.Println("\n=== Simple Model ===")
			fmt.Println(k)
			runModelChecking(k, reader)

		case "4":
			if !hasOpenAI {
				fmt.Println("OpenAI API key not set!")
				continue
			}
			fmt.Print("\nDescribe the system in English: ")
			description, _ := reader.ReadString('\n')
			description = strings.TrimSpace(description)

			fmt.Println("Generating model from description...")
			client := NewOpenAIClient()
			model, err := client.GenerateModel(description)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			fmt.Println("\n=== Generated Model ===")
			fmt.Println(model)
			runModelChecking(model, reader)

		case "5":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid option")
		}
	}
}

func runModelChecking(k *KripkeStructure, reader *bufio.Reader) {
	mc := NewModelChecker(k)
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""

	fmt.Println("\n=== Model Checking ===")
	fmt.Println("Available propositions:", getAvailablePropositions(k))
	fmt.Println()

	// Run some example queries
	runExampleQueries(mc)

	fmt.Println("\nCustom queries:")
	fmt.Println("1. Check predefined CTL formulas")
	if hasOpenAI {
		fmt.Println("2. Enter query in English (OpenAI)")
	}
	fmt.Println("3. Return to main menu")
	fmt.Print("\nSelect option: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		checkPredefinedFormulas(mc, k, reader)
	case "2":
		if !hasOpenAI {
			fmt.Println("OpenAI API key not set!")
			return
		}
		checkEnglishQuery(mc, reader)
	case "3":
		return
	}
}

func getAvailablePropositions(k *KripkeStructure) []Proposition {
	propSet := make(map[Proposition]bool)
	for _, labels := range k.Labeling {
		for _, prop := range labels {
			propSet[prop] = true
		}
	}

	props := make([]Proposition, 0, len(propSet))
	for prop := range propSet {
		props = append(props, prop)
	}
	return props
}

func runExampleQueries(mc *ModelChecker) {
	props := getAvailablePropositions(mc.Structure)
	if len(props) == 0 {
		fmt.Println("No propositions available for queries")
		return
	}

	// Example: Check if first proposition holds in initial state
	p := props[0]
	formula := AtomicProp{p}
	result := mc.Holds(formula)
	fmt.Printf("Initial state satisfies '%s': %v\n", p, result)

	// Example: Check EF (eventually) for first proposition
	efFormula := EF{AtomicProp{p}}
	result = mc.Holds(efFormula)
	fmt.Printf("EF %s (eventually %s): %v\n", p, p, result)

	// Example: Check AG (always) for first proposition if we have more than one
	if len(props) > 1 {
		q := props[1]
		agFormula := AG{AtomicProp{q}}
		result = mc.Holds(agFormula)
		fmt.Printf("AG %s (always %s): %v\n", q, q, result)
	}
}

func checkPredefinedFormulas(mc *ModelChecker, k *KripkeStructure, reader *bufio.Reader) {
	props := getAvailablePropositions(k)
	if len(props) == 0 {
		fmt.Println("No propositions available")
		return
	}

	p := props[0]
	var q Proposition
	if len(props) > 1 {
		q = props[1]
	} else {
		q = p
	}

	formulas := []struct {
		name    string
		formula CTLFormula
	}{
		{"p (atomic)", AtomicProp{p}},
		{"EX p (exists next p)", EX{AtomicProp{p}}},
		{"AX p (all next p)", AX{AtomicProp{p}}},
		{"EF p (eventually p)", EF{AtomicProp{p}}},
		{"AF p (always eventually p)", AF{AtomicProp{p}}},
		{"EG p (exists always p)", EG{AtomicProp{p}}},
		{"AG p (always p)", AG{AtomicProp{p}}},
	}

	if len(props) > 1 {
		formulas = append(formulas,
			struct {
				name    string
				formula CTLFormula
			}{"E[p U q] (p until q)", EU{AtomicProp{p}, AtomicProp{q}}},
		)
	}

	fmt.Println("\nChecking formulas in initial state:")
	for _, f := range formulas {
		result := mc.Holds(f.formula)
		fmt.Printf("  %s: %v\n", f.name, result)
	}
}

func checkEnglishQuery(mc *ModelChecker, reader *bufio.Reader) {
	fmt.Print("\nEnter query in English: ")
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	fmt.Println("Converting to CTL formula...")
	client := NewOpenAIClient()
	formulaStr, err := client.GenerateCTLFormula(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Generated CTL formula: %s\n", formulaStr)
	fmt.Println("(Note: Parsing arbitrary CTL formulas requires a parser implementation)")
	fmt.Println("This would be the next step for full functionality")
}
