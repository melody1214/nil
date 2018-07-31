package matrix

import (
	"reflect"
	"testing"
)

func TestFindEncodingMatrixByIndex(t *testing.T) {
	testCases := []struct {
		index    int
		expected EncodingMatrix
	}{
		{
			1,
			EncodingMatrix{
				ID: ID(106),
				Matrix: [][]byte{
					[]byte{1, 0, 0, 0},
					[]byte{0, 1, 0, 0},
					[]byte{0, 0, 1, 0},
					[]byte{0, 0, 0, 1},
					[]byte{133, 125, 168, 58},
					[]byte{125, 133, 58, 168},
					[]byte{168, 58, 133, 125},
					[]byte{58, 168, 125, 133},
				},
			},
		},
		{
			49,
			EncodingMatrix{
				ID: ID(92),
				Matrix: [][]byte{
					[]byte{1, 0, 0, 0},
					[]byte{0, 1, 0, 0},
					[]byte{0, 0, 1, 0},
					[]byte{0, 0, 0, 1},
					[]byte{61, 170, 93, 150},
					[]byte{170, 61, 150, 93},
					[]byte{93, 150, 61, 170},
					[]byte{150, 93, 170, 61},
				},
			},
		},
	}

	for _, c := range testCases {
		m, err := FindEncodingMatrixByIndex(c.index)
		if err != nil {
			t.Errorf("met unexpected error: %v", err)
			continue
		}

		if !reflect.DeepEqual(m, &c.expected) {
			t.Errorf("expected encoding matrix \n%s, got \n%s", c.expected.String(), m.String())
		}
	}
}

func TestFindEncodingMatrixByID(t *testing.T) {
	testCases := []struct {
		id       ID
		expected EncodingMatrix
	}{
		{
			ID(106),
			EncodingMatrix{
				ID: ID(106),
				Matrix: [][]byte{
					[]byte{1, 0, 0, 0},
					[]byte{0, 1, 0, 0},
					[]byte{0, 0, 1, 0},
					[]byte{0, 0, 0, 1},
					[]byte{133, 125, 168, 58},
					[]byte{125, 133, 58, 168},
					[]byte{168, 58, 133, 125},
					[]byte{58, 168, 125, 133},
				},
			},
		},
		{
			ID(92),
			EncodingMatrix{
				ID: ID(92),
				Matrix: [][]byte{
					[]byte{1, 0, 0, 0},
					[]byte{0, 1, 0, 0},
					[]byte{0, 0, 1, 0},
					[]byte{0, 0, 0, 1},
					[]byte{61, 170, 93, 150},
					[]byte{170, 61, 150, 93},
					[]byte{93, 150, 61, 170},
					[]byte{150, 93, 170, 61},
				},
			},
		},
	}

	for _, c := range testCases {
		m, err := FindEncodingMatrixByID(c.id)
		if err != nil {
			t.Errorf("met unexpected error: %v", err)
			continue
		}

		if !reflect.DeepEqual(m, &c.expected) {
			t.Errorf("expected encoding matrix \n%s, got \n%s", c.expected.String(), m.String())
		}
	}
}

func TestDuplicateID(t *testing.T) {
	for i, id := range cauchyID {
		for j, dup := range cauchyID {
			if i == j {
				continue
			}
			if id == dup {
				t.Errorf("duplicate id exist: %d of %d'th index and %d of %d'th index", id, i, dup, j)
			}
		}
	}
}
