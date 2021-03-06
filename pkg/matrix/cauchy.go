package matrix

import "errors"

var (
	// ErrIndexOutOfBound is used when the given index is out of bound of cauchy matrix or id table.
	ErrIndexOutOfBound = errors.New("index is out of bound")

	// ErrNoSuchID is used when the given id of encoding matrix is not exist.
	ErrNoSuchID = errors.New("no such encoding matrix with the given id")
)

const maxEncodingMatrixSize = 50

var cauchyTable = [4][200]byte{
	{210, 247, 98, 90, 133, 125, 168, 58, 18, 89, 165, 53, 101, 184, 163, 158, 16, 115, 118, 120, 153, 10, 25, 145, 41, 113, 200, 246, 249, 67, 215, 214, 250, 116, 243, 180, 109, 33, 178, 106, 20, 63, 230, 240, 134, 177, 226, 241, 66, 212, 232, 117, 127, 255, 126, 253, 227, 231, 181, 234, 3, 143, 211, 201, 73, 49, 39, 45, 83, 105, 2, 245, 27, 84, 161, 29, 124, 204, 228, 176, 11, 220, 189, 148, 172, 9, 199, 162, 24, 223, 68, 79, 155, 188, 15, 92, 206, 59, 13, 60, 156, 8, 190, 183, 28, 130, 159, 198, 52, 194, 70, 5, 197, 100, 7, 123, 149, 154, 174, 182, 135, 229, 238, 107, 235, 242, 191, 175, 48, 136, 43, 30, 22, 103, 69, 147, 54, 95, 248, 213, 146, 78, 166, 4, 19, 193, 203, 99, 151, 14, 55, 65, 56, 35, 104, 140, 129, 26, 37, 97, 82, 141, 239, 179, 32, 236, 47, 50, 36, 87, 202, 91, 185, 196, 23, 77, 219, 119, 6, 187, 132, 205, 254, 252, 40, 209, 17, 217, 233, 251, 218, 121, 173, 157, 221, 152, 61, 170, 93, 150},
	{247, 210, 90, 98, 125, 133, 58, 168, 89, 18, 53, 165, 184, 101, 158, 163, 115, 16, 120, 118, 10, 153, 145, 25, 113, 41, 246, 200, 67, 249, 214, 215, 116, 250, 180, 243, 33, 109, 106, 178, 63, 20, 240, 230, 177, 134, 241, 226, 212, 66, 117, 232, 255, 127, 253, 126, 231, 227, 234, 181, 143, 3, 201, 211, 49, 73, 45, 39, 105, 83, 245, 2, 84, 27, 29, 161, 204, 124, 176, 228, 220, 11, 148, 189, 9, 172, 162, 199, 223, 24, 79, 68, 188, 155, 92, 15, 59, 206, 60, 13, 8, 156, 183, 190, 130, 28, 198, 159, 194, 52, 5, 70, 100, 197, 123, 7, 154, 149, 182, 174, 229, 135, 107, 238, 242, 235, 175, 191, 136, 48, 30, 43, 103, 22, 147, 69, 95, 54, 213, 248, 78, 146, 4, 166, 193, 19, 99, 203, 14, 151, 65, 55, 35, 56, 140, 104, 26, 129, 97, 37, 141, 82, 179, 239, 236, 32, 50, 47, 87, 36, 91, 202, 196, 185, 77, 23, 119, 219, 187, 6, 205, 132, 252, 254, 209, 40, 217, 17, 251, 233, 121, 218, 157, 173, 152, 221, 170, 61, 150, 93},
	{98, 90, 210, 247, 168, 58, 133, 125, 165, 53, 18, 89, 163, 158, 101, 184, 118, 120, 16, 115, 25, 145, 153, 10, 200, 246, 41, 113, 215, 214, 249, 67, 243, 180, 250, 116, 178, 106, 109, 33, 230, 240, 20, 63, 226, 241, 134, 177, 232, 117, 66, 212, 126, 253, 127, 255, 181, 234, 227, 231, 211, 201, 3, 143, 39, 45, 73, 49, 2, 245, 83, 105, 161, 29, 27, 84, 228, 176, 124, 204, 189, 148, 11, 220, 199, 162, 172, 9, 68, 79, 24, 223, 15, 92, 155, 188, 13, 60, 206, 59, 190, 183, 156, 8, 159, 198, 28, 130, 70, 5, 52, 194, 7, 123, 197, 100, 174, 182, 149, 154, 238, 107, 135, 229, 191, 175, 235, 242, 43, 30, 48, 136, 69, 147, 22, 103, 248, 213, 54, 95, 166, 4, 146, 78, 203, 99, 19, 193, 55, 65, 151, 14, 104, 140, 56, 35, 37, 97, 129, 26, 239, 179, 82, 141, 47, 50, 32, 236, 202, 91, 36, 87, 23, 77, 185, 196, 6, 187, 219, 119, 254, 252, 132, 205, 17, 217, 40, 209, 218, 121, 233, 251, 221, 152, 173, 157, 93, 150, 61, 170},
	{90, 98, 247, 210, 58, 168, 125, 133, 53, 165, 89, 18, 158, 163, 184, 101, 120, 118, 115, 16, 145, 25, 10, 153, 246, 200, 113, 41, 214, 215, 67, 249, 180, 243, 116, 250, 106, 178, 33, 109, 240, 230, 63, 20, 241, 226, 177, 134, 117, 232, 212, 66, 253, 126, 255, 127, 234, 181, 231, 227, 201, 211, 143, 3, 45, 39, 49, 73, 245, 2, 105, 83, 29, 161, 84, 27, 176, 228, 204, 124, 148, 189, 220, 11, 162, 199, 9, 172, 79, 68, 223, 24, 92, 15, 188, 155, 60, 13, 59, 206, 183, 190, 8, 156, 198, 159, 130, 28, 5, 70, 194, 52, 123, 7, 100, 197, 182, 174, 154, 149, 107, 238, 229, 135, 175, 191, 242, 235, 30, 43, 136, 48, 147, 69, 103, 22, 213, 248, 95, 54, 4, 166, 78, 146, 99, 203, 193, 19, 65, 55, 14, 151, 140, 104, 35, 56, 97, 37, 26, 129, 179, 239, 141, 82, 50, 47, 236, 32, 91, 202, 87, 36, 77, 23, 196, 185, 187, 6, 119, 219, 252, 254, 205, 132, 217, 17, 209, 40, 121, 218, 251, 233, 152, 221, 157, 173, 150, 93, 170, 61},
}

var cauchyID [maxEncodingMatrixSize]ID

// FindEncodingMatrixByIndex returns a encoding matrix with the given index.
func FindEncodingMatrixByIndex(index int) (*EncodingMatrix, error) {
	if maxEncodingMatrixSize <= index {
		return nil, ErrIndexOutOfBound
	}

	m := &EncodingMatrix{
		ID:     ID(cauchyID[index]),
		Matrix: make([][]byte, defaultMatrixSize+defaultMatrixSize),
	}
	for i := range m.Matrix {
		m.Matrix[i] = make([]byte, defaultMatrixSize)
	}

	// Make identity matrix on top.
	for i := 0; i < defaultMatrixSize; i++ {
		m.Matrix[i][i] = 1
	}

	for row := 0; row < defaultMatrixSize; row++ {
		for col := 0; col < defaultMatrixSize; col++ {
			m.Matrix[defaultMatrixSize+row][col] = cauchyTable[row][(index*defaultMatrixSize)+col]
		}
	}

	return m, nil
}

// FindEncodingMatrixByID returns a encoding matrix with the given index.
func FindEncodingMatrixByID(id ID) (*EncodingMatrix, error) {
	for i, cid := range cauchyID {
		if cid == id {
			return FindEncodingMatrixByIndex(i)
		}
	}
	return nil, ErrNoSuchID
}

func init() {
	for i := 0; i < len(cauchyID); i++ {
		id := byte(0)
		for j := 0; j < defaultMatrixSize; j++ {
			id = id ^ cauchyTable[0][(i*defaultMatrixSize)+j]
		}

		cauchyID[i] = ID(id)
	}
}
