package swim

import (
	"testing"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

func TestSwimAlgorithm(t *testing.T) {
	testCases := []struct {
		old      *swimpb.Member
		new      *swimpb.Member
		solution bool
	}{
		// Case 0.
		{
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 1},
			true,
		},
		// Case 1.
		{
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 1},
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 0},
			false,
		},
		// Case 2.
		{
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 1},
			true,
		},
		// Case 3.
		{
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 0},
			false,
		},
		// Case 4.
		{
			&swimpb.Member{Status: swimpb.Status_FAULTY, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 1},
			false,
		},
		// Case 5.
		{
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			true,
		},
		// Case 6.
		{
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 1},
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			false,
		},
		// Case 7.
		{
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			false,
		},
		// Case 8.
		{
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 1},
			true,
		},
		// Case 9.
		{
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 1},
			&swimpb.Member{Status: swimpb.Status_FAULTY, Incarnation: 0},
			true,
		},
		// Case 10.
		{
			&swimpb.Member{Status: swimpb.Status_ALIVE, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_FAULTY, Incarnation: 1},
			true,
		},
		// Case 11.
		{
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 1},
			&swimpb.Member{Status: swimpb.Status_FAULTY, Incarnation: 0},
			true,
		},
		// Case 12.
		{
			&swimpb.Member{Status: swimpb.Status_SUSPECT, Incarnation: 0},
			&swimpb.Member{Status: swimpb.Status_FAULTY, Incarnation: 1},
			true,
		},
	}

	for i, c := range testCases {
		if answer := compare(c.old, c.new); answer != c.solution {
			t.Errorf("test-case(%d): expected answer %t, got %t", i, c.solution, answer)
		}
	}
}
