package domain

import (
	"fmt"
	"testing"
	"time"
)

var e1 Entrant = Entrant{12394650, "LG | Tweek", 3, 9, true}
var e2 Entrant = Entrant{12687800, "Zomba", 20, 8, false}
var s1 []Selection = []Selection{{e1, &Character{1279, "Diddy Kong"}}, {e2, &Character{1323, "R.O.B."}}}
var s2 []Selection = []Selection{{e1, &Character{1777, "Sephiroth"}}, {e2, &Character{1323, "R.O.B."}}}

type setTestCase struct {
	set                                                         *Set
	expected, expectedWinnersSelection, expectedLosersSelection string
	expectedIsDQAndOut, expectedIsNotable                       bool
}

var setTestCases []setTestCase = []setTestCase{
	{
		NewSet(
			"60482457",
			"Zomba 3 - LG | Tweek 0",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			nil,
			int(time.Now().UnixMilli()),
		),
		"3-0",
		"",
		"",
		false,
		false,
	},
	{
		NewSet(
			"60482457",
			"Zomba 0 - LG | Tweek 3",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			nil,
			int(time.Now().UnixMilli()),
		),
		"3-0",
		"",
		"",
		false,
		false,
	},
	{
		NewSet(
			"60482457",
			"Zomba 0 - LG | Tweek 3",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			&[]Game{{16955184, 12687800, s1}, {16955185, 12687800, s2}, {16955186, 12687800, s2}},
			int(time.Now().UnixMilli()),
		),
		"3-0",
		"R.O.B.",
		"Diddy Kong, Sephiroth",
		false,
		false,
	},
	{
		NewSet(
			"60482457",
			"Zomba 0 - LG | Tweek 3",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			&[]Game{{16955184, 12687800, s1}, {16955185, 12687800, s1}, {16955186, 12394650, s2}, {16955187, 12687800, s1}},
			int(time.Now().UnixMilli()),
		),
		"3-1",
		"R.O.B.",
		"Diddy Kong, Sephiroth",
		false,
		false,
	},
	{
		NewSet(
			"60482457",
			"DQ",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			&[]Game{{16955184, 12687800, s1}, {16955185, 12687800, s1}, {16955186, 12394650, s2}, {16955187, 12687800, s1}},
			int(time.Now().UnixMilli()),
		),
		"DQ",
		"R.O.B.",
		"Diddy Kong, Sephiroth",
		true,
		false,
	},
	{
		NewSet(
			"60482457",
			"Zomba 1 - LG | Tweek 2",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			nil,
			int(time.Now().UnixMilli()),
		),
		"2-1",
		"",
		"",
		false,
		true,
	},
	{
		NewSet(
			"60482457",
			"Zomba 0 - LG | Tweek 3",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			&[]Game{{16955184, 12687800, s1}, {16955185, 12687800, s1}, {169551846, 12394650, s2}},
			int(time.Now().UnixMilli()),
		),
		"3-0",
		"R.O.B.",
		"Diddy Kong, Sephiroth",
		false,
		false,
	},
	{
		NewSet(
			"60482457",
			"Zomba 2 - LG | Tweek 3",
			nil,
			5,
			-6,
			9,
			12687800,
			[]Entrant{e1, e2},
			nil,
			int(time.Now().UnixMilli()),
		),
		"3-2",
		"",
		"",
		false,
		true,
	},
}

func TestSet(t *testing.T) {
	for idx, test := range setTestCases {
		t.Run(fmt.Sprintf("Test case %v", idx+1), func(t *testing.T) {
			result := *test.set.Score
			expected := test.expected
			if result != expected {
				t.Errorf("Scores failed. Result %s, expected %s", result, expected)
			}
			resultWinnersSelection := test.set.GetWinnerCharacterSelections()
			expectedWinnersSelection := test.expectedWinnersSelection
			if resultWinnersSelection != expectedWinnersSelection {
				t.Errorf("Winners character selections failed. Result %s, expected %s", resultWinnersSelection, expectedWinnersSelection)
			}
			resultLosersSelection := test.set.GetLoserCharacterSelections()
			expectedLosersSelection := test.expectedLosersSelection
			if resultLosersSelection != expectedLosersSelection {
				t.Errorf("Losers character selections failed. Result %s, expected %s", resultLosersSelection, expectedLosersSelection)
			}
			resultIsDQAndOut := test.set.IsDQAndOut()
			expectedIsDQAndOut := test.expectedIsDQAndOut
			if resultIsDQAndOut != expectedIsDQAndOut {
				t.Errorf("DQ and out failed. Result %t, expected %t", resultIsDQAndOut, expectedIsDQAndOut)
			}
			resultIsNotable := test.set.IsNotable()
			expectedIsNotable := test.expectedIsNotable
			if resultIsNotable != expectedIsNotable {
				t.Errorf("Notable failed. Result %t, expected %t", resultIsNotable, expectedIsNotable)
			}
		})
	}
}
