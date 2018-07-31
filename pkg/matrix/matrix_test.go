package matrix

import "testing"
import "reflect"

func TestMerge(t *testing.T) {
	testCases := []struct {
		matrices []EncodingMatrix
		expected EncodingMatrix
	}{
		{
			[]EncodingMatrix{
				EncodingMatrix{
					ID: 0x2,
					Matrix: [][]byte{
						[]byte{1, 0, 0, 0},
						[]byte{0, 1, 0, 0},
						[]byte{0, 0, 1, 0},
						[]byte{0, 0, 0, 1},
						[]byte{149, 154, 174, 182},
						[]byte{154, 149, 182, 174},
						[]byte{174, 182, 149, 154},
						[]byte{182, 174, 154, 149},
					},
				},
				EncodingMatrix{
					ID: 0x1,
					Matrix: [][]byte{
						[]byte{1, 0, 0, 0},
						[]byte{0, 1, 0, 0},
						[]byte{0, 0, 1, 0},
						[]byte{0, 0, 0, 1},
						[]byte{219, 119, 6, 187},
						[]byte{119, 219, 187, 6},
						[]byte{6, 187, 219, 119},
						[]byte{187, 6, 119, 219},
					},
				},
				EncodingMatrix{
					ID: 0x3,
					Matrix: [][]byte{
						[]byte{1, 0, 0, 0},
						[]byte{0, 1, 0, 0},
						[]byte{0, 0, 1, 0},
						[]byte{0, 0, 0, 1},
						[]byte{101, 184, 163, 158},
						[]byte{184, 101, 158, 163},
						[]byte{163, 158, 101, 184},
						[]byte{158, 163, 184, 101},
					},
				},
			},
			EncodingMatrix{
				Matrix: [][]byte{
					[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					[]byte{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					[]byte{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					[]byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
					[]byte{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
					[]byte{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
					[]byte{0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
					[]byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
					[]byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
					[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
					[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
					[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
					[]byte{219, 119, 6, 187, 149, 154, 174, 182, 101, 184, 163, 158},
					[]byte{119, 219, 187, 6, 154, 149, 182, 174, 184, 101, 158, 163},
					[]byte{6, 187, 219, 119, 174, 182, 149, 154, 163, 158, 101, 184},
					[]byte{187, 6, 119, 219, 182, 174, 154, 149, 158, 163, 184, 101},
				},
			},
		},
	}

	for _, c := range testCases {
		merged := Merge(c.matrices...)
		if !reflect.DeepEqual(merged, c.expected) {
			t.Errorf("expected merged matrix \n%s, got \n%s", c.expected.String(), merged.String())
		}
	}
}
