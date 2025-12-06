
package purple

import "github.com/rfielding/kripke-ctl/kripke"

type PurpleModel struct{}

func (PurpleModel) Name() string { return "purple" }

func (PurpleModel) OriginalText() string {
    return `Scenario: Japanese PURPLE diplomatic cipher keyspace collapse.

We model the attacker's progress in shrinking the key hypothesis space
as more traffic is observed. Because the cipher splits the alphabet
into independent vowel and consonant channels, the effective keyspace
collapses much faster than for a well-designed rotor machine.`
}

func (PurpleModel) BuildGraph() (*kripke.SimpleGraph, kripke.NodeID) {
    return kripke.BuildPurpleGraph()
}

func (PurpleModel) CTLFormulas() []kripke.CTLSpec {
    return []kripke.CTLSpec{
        {
            Name:        "AF attackerKnowsKey",
            Description: "Eventually the attacker uniquely knows the key.",
            Formula:     "AF attackerKnowsKey",
        },
        {
            Name:        "AG !attackerKnowsKey",
            Description: "Key remains forever unknown (desired but false).",
            Formula:     "AG !attackerKnowsKey",
        },
    }
}

func (PurpleModel) Counters() []kripke.CounterSpec {
    return []kripke.CounterSpec{
        {
            Name:        "keySpaceSize",
            Description: "Size of the remaining key hypothesis set in a simulation.",
        },
    }
}
