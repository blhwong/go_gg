package mapper

import "testing"

type testCase struct {
	name     string
	input    int
	expected string
}

var testCases []testCase = []testCase{
	{"1 is 1st", 1, "1st"},
	{"22 is 22nd", 22, "22nd"},
	{"33 is 33rd", 33, "33rd"},
	{"44 is 44th", 44, "44th"},
	{"11 is 11th", 11, "11th"},
}

func TestGetOrdinal(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := getOrdinal(tc.input)
			if res != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, res)
			}
		})
	}
}
