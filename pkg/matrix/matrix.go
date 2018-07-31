package matrix

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// EncodingMatrix represents an encoding matrix.
type EncodingMatrix struct {
	ID     ID
	Matrix [][]byte
}

func (m EncodingMatrix) String() string {
	idOut := strconv.Itoa(int(m.ID.Byte()))
	rowOut := make([]string, 0, len(m.Matrix))
	for _, row := range m.Matrix {
		colOut := make([]string, 0, len(row))
		for _, col := range row {
			colOut = append(colOut, fmt.Sprintf("%3s", strconv.Itoa(int(col))))
		}
		rowOut = append(rowOut, "["+strings.Join(colOut, ", ")+"]")
	}
	return "<ID: " + idOut + ">" + "\n" + strings.Join(rowOut, "\n")
}

// ID represents an ID of encoding matrix.
type ID byte

// Byte returns a byte of ID.
func (id ID) Byte() byte {
	return byte(id)
}

const (
	// default encoding matrix is 8x4
	defaultMatrixSize = 4
)

// Merge combines multiple encoding matrices into one.
func Merge(matrices ...EncodingMatrix) EncodingMatrix {
	// Sort matrices by id.
	sort.Slice(matrices, func(i, j int) bool {
		return matrices[i].ID.Byte() < matrices[j].ID.Byte()
	})

	var merged EncodingMatrix
	colSize := defaultMatrixSize * len(matrices)
	rowSize := (defaultMatrixSize * len(matrices)) + defaultMatrixSize

	merged.Matrix = make([][]byte, rowSize)
	for i := range merged.Matrix {
		merged.Matrix[i] = make([]byte, colSize)
	}

	// Make identity matrix on top.
	for i := 0; i < colSize; i++ {
		merged.Matrix[i][i] = 1
	}

	for i, m := range matrices {
		for r := 0; r < defaultMatrixSize; r++ {
			for c := 0; c < defaultMatrixSize; c++ {
				merged.Matrix[colSize+r][(defaultMatrixSize*i)+c] = m.Matrix[defaultMatrixSize+r][c]
			}
		}
	}

	return merged
}
