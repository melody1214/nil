package swim

import (
	"testing"
)

func TestSwimAlgorithm(t *testing.T) {
	testCases := []struct {
		old      Member
		new      Member
		solution bool
	}{
		// Case 0.
		{
			Member{Status: Alive, Incarnation: 0},
			Member{Status: Alive, Incarnation: 1},
			true,
		},
		// Case 1.
		{
			Member{Status: Alive, Incarnation: 1},
			Member{Status: Alive, Incarnation: 0},
			false,
		},
		// Case 2.
		{
			Member{Status: Suspect, Incarnation: 0},
			Member{Status: Alive, Incarnation: 1},
			true,
		},
		// Case 3.
		{
			Member{Status: Suspect, Incarnation: 0},
			Member{Status: Alive, Incarnation: 0},
			false,
		},
		// Case 4.
		{
			Member{Status: Faulty, Incarnation: 0},
			Member{Status: Alive, Incarnation: 1},
			false,
		},
		// Case 5.
		{
			Member{Status: Alive, Incarnation: 0},
			Member{Status: Suspect, Incarnation: 0},
			true,
		},
		// Case 6.
		{
			Member{Status: Alive, Incarnation: 1},
			Member{Status: Suspect, Incarnation: 0},
			false,
		},
		// Case 7.
		{
			Member{Status: Suspect, Incarnation: 0},
			Member{Status: Suspect, Incarnation: 0},
			false,
		},
		// Case 8.
		{
			Member{Status: Suspect, Incarnation: 0},
			Member{Status: Suspect, Incarnation: 1},
			true,
		},
		// Case 9.
		{
			Member{Status: Alive, Incarnation: 1},
			Member{Status: Faulty, Incarnation: 0},
			true,
		},
		// Case 10.
		{
			Member{Status: Alive, Incarnation: 0},
			Member{Status: Faulty, Incarnation: 1},
			true,
		},
		// Case 11.
		{
			Member{Status: Suspect, Incarnation: 1},
			Member{Status: Faulty, Incarnation: 0},
			true,
		},
		// Case 12.
		{
			Member{Status: Suspect, Incarnation: 0},
			Member{Status: Faulty, Incarnation: 1},
			true,
		},
	}

	for i, c := range testCases {
		if answer := compare(c.old, c.new); answer != c.solution {
			t.Errorf("test-case(%d): expected answer %t, got %t", i, c.solution, answer)
		}
	}
}
