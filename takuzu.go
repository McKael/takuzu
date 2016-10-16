package takuzu

import (
	"bytes"
	"fmt"
	"math"

	"github.com/pkg/errors"
)

// Cell is a single cell of a Takuzu game board
type Cell struct {
	Defined bool
	Value   int
}

// Takuzu is a Takuzu game board (Size x Size)
type Takuzu struct {
	Size  int
	Board [][]Cell
}

// New creates a new Takuzu board
func New(size int) Takuzu {
	t := Takuzu{Size: size}
	t.Board = make([][]Cell, size)
	for l := range t.Board {
		t.Board[l] = make([]Cell, size)
	}
	return t
}

// NewFromString creates a new Takuzu board from a string definition
func NewFromString(s string) (*Takuzu, error) {
	l := len(s)
	if l < 4 {
		return nil, errors.New("bad string length")
	}

	size := int(math.Sqrt(float64(l)))
	if size*size != l {
		return nil, errors.New("bad string length")
	}

	i := 0
	t := New(size)

	for line := 0; line < size; line++ {
		for col := 0; col < size; col++ {
			switch s[i] {
			case '0', 'O':
				t.Board[line][col].Defined = true
				t.Board[line][col].Value = 0
			case '1', 'I':
				t.Board[line][col].Defined = true
				t.Board[line][col].Value = 1
			case '.':
			default:
				return nil, errors.New("invalid char in string")
			}
			i++
		}
	}
	return &t, nil
}

// ToString converts a takuzu board to its string representation
func (b Takuzu) ToString() string {
	var sbuf bytes.Buffer
	for line := 0; line < b.Size; line++ {
		for col := 0; col < b.Size; col++ {
			if b.Board[line][col].Defined {
				sbuf.WriteString(fmt.Sprintf("%d", b.Board[line][col].Value))
				continue
			}
			sbuf.WriteByte('.')
		}
	}
	return sbuf.String()
}

// DumpString writes the content of the board as a string
func (b Takuzu) DumpString() {
	fmt.Println(b.ToString())
}

// Clone returns a copy of the Takuzu board
func (b Takuzu) Clone() Takuzu {
	c := New(b.Size)
	for line := range b.Board {
		for col := range b.Board[line] {
			c.Board[line][col] = b.Board[line][col]
		}
	}
	return c
}

// Copy copies a Takuzu board to another existing board
func Copy(src, dst *Takuzu) error {
	if src.Size != dst.Size {
		return errors.New("sizes do not match")
	}
	for line := range src.Board {
		for col := range src.Board[line] {
			dst.Board[line][col] = src.Board[line][col]
		}
	}
	return nil
}

// BoardsMatch compares a Takuzu board to another, optionally ignoring
// empty cells.  Returns true if the two boards match.
func BoardsMatch(t1, t2 *Takuzu, ignoreUndefined bool) (match bool, line, col int) {
	match = true

	if t1 == nil || t2 == nil {
		line, col = -1, -1
		match = false
		return
	}

	if t1.Size != t2.Size {
		line, col = -1, -1
		match = false
		return
	}

	for line = range t1.Board {
		for col = range t1.Board[line] {
			if !t1.Board[line][col].Defined || !t2.Board[line][col].Defined {
				// At least one of the cells is empty
				if ignoreUndefined ||
					!(t1.Board[line][col].Defined || t2.Board[line][col].Defined) {
					// Both cells are empty or we ignore empty cells
					continue
				}
				match = false
				return
			}
			// Both cells are defined
			if t1.Board[line][col].Value != t2.Board[line][col].Value {
				match = false
				return
			}
		}
	}

	line, col = -1, -1
	return
}

// Set sets the value of the cell; a value -1 will set the cell as undefined
func (c *Cell) Set(value int) {
	if value != 0 && value != 1 {
		c.Defined = false
		return
	}
	c.Defined = true
	c.Value = value
}

// Set sets the value of a specific cell
// A value -1 will undefine the cell
func (b Takuzu) Set(l, c, value int) {
	if value != 0 && value != 1 {
		b.Board[l][c].Defined = false
		return
	}
	b.Board[l][c].Defined = true
	b.Board[l][c].Value = value
}

// GetLine returns a slice of cells containing the ith line of the board
func (b Takuzu) GetLine(i int) []Cell {
	return b.Board[i]
}

// GetColumn returns a slice of cells containing the ith column of the board
func (b Takuzu) GetColumn(i int) []Cell {
	c := make([]Cell, b.Size)
	for l := range b.Board {
		c[l] = b.Board[l][i]
	}
	return c
}

// GetLinePointers returns a slice of pointers to the cells of the ith line of the board
func (b Takuzu) GetLinePointers(i int) []*Cell {
	r := make([]*Cell, b.Size)
	for l := range b.Board[i] {
		r[l] = &b.Board[i][l]
	}
	return r
}

// GetColumnPointers returns a slice of pointers to the cells of the ith column of the board
func (b Takuzu) GetColumnPointers(i int) []*Cell {
	r := make([]*Cell, b.Size)
	for l := range b.Board {
		r[l] = &b.Board[l][i]
	}
	return r
}

// FillLineColumn add missing 0s or 1s if all 1s or 0s are there.
// Note: This method can update b.
func (b Takuzu) FillLineColumn(l, c int) {
	fillRange := func(r []*Cell) {
		size := len(r)
		var notFull bool
		var n [2]int
		for x := 0; x < size; x++ {
			if r[x].Defined {
				n[r[x].Value]++
			} else {
				notFull = true
			}
		}
		if !notFull {
			return
		}
		if n[0] == size/2 {
			// Let's fill the 1s
			for _, x := range r {
				if !x.Defined {
					x.Defined = true
					x.Value = 1
				}
			}
		} else if n[1] == size/2 {
			// Let's fill the 0s
			for _, x := range r {
				if !x.Defined {
					x.Defined = true
					x.Value = 0
				}
			}
		}
	}

	var cells []*Cell

	// Fill line
	cells = b.GetLinePointers(l)
	fillRange(cells)
	// Fill column
	cells = b.GetColumnPointers(c)
	fillRange(cells)
}

// DumpBoard displays the Takuzu board
func (b Takuzu) DumpBoard() {
	fmt.Println()
	for i := range b.Board {
		dumpRange(b.Board[i])
	}
}

// CheckLine returns an error if the line i fails validation
func (b Takuzu) CheckLine(i int) error {
	_, err := checkRange(b.GetLine(i))
	return err
}

// CheckColumn returns an error if the column i fails validation
func (b Takuzu) CheckColumn(i int) error {
	_, err := checkRange(b.GetColumn(i))
	return err
}

// Validate checks a whole board for errors (not completeness)
// Returns true if all cells are defined.
func (b Takuzu) Validate() (bool, error) {
	finished := true

	computeVal := func(cells []Cell) (val int) {
		for i := 0; i < len(cells); i++ {
			val += cells[i].Value * 1 << uint(i)
		}
		return
	}

	lineVals := make(map[int]bool)
	colVals := make(map[int]bool)

	for i := 0; i < b.Size; i++ {
		var d []Cell
		var full bool
		var err error

		// Let's check line i
		d = b.GetLine(i)
		full, err = checkRange(d)
		if err != nil {
			return false, errors.Wrapf(err, "line %d", i)
		}
		if full {
			hv := computeVal(d)
			if lineVals[hv] {
				return false, fmt.Errorf("duplicate lines (%d)", i)
			}
			lineVals[hv] = true
		} else {
			finished = false
		}

		// Let's check column i
		d = b.GetColumn(i)
		full, err = checkRange(d)
		if err != nil {
			return false, errors.Wrapf(err, "column %d", i)
		}
		if full {
			hv := computeVal(d)
			if colVals[hv] {
				return false, fmt.Errorf("duplicate columns (%d)", i)
			}
			colVals[hv] = true
		} else {
			finished = false
		}
	}
	return finished, nil
}

func dumpRange(cells []Cell) {
	for _, c := range cells {
		if !c.Defined {
			fmt.Printf(". ")
			continue
		}
		fmt.Printf("%d ", c.Value)
	}
	fmt.Println()
}

// checkRange returns true if the range is completely defined, and an error
// if it doesn't follow the rules for a takuzu line or column
// Note that the boolean might be invalid if the error is not nil.
func checkRange(cells []Cell) (bool, error) {
	full := true
	size := len(cells)
	counters := []int{0, 0}

	var prevCell Cell
	var prevCellCount int

	for _, c := range cells {
		if !c.Defined {
			full = false
			prevCell.Defined = false
			prevCellCount = 0
			continue
		}
		counters[c.Value]++
		if prevCellCount == 0 {
			prevCellCount = 1
		} else {
			if c.Value == prevCell.Value {
				prevCellCount++
				if prevCellCount > 2 {
					return full, errors.Errorf("3+ same values %d", c.Value)
				}
			} else {
				prevCellCount = 1
			}

		}
		prevCell = c
	}
	if counters[0] > size/2 {
		return full, errors.Errorf("too many zeroes")
	}
	if counters[1] > size/2 {
		return full, errors.Errorf("too many ones")
	}
	return full, nil
}

func checkRangeCounts(cells []Cell) (full bool, n0, n1 int) {
	counters := []int{0, 0}
	full = true

	for _, c := range cells {
		if c.Defined {
			counters[c.Value]++
		} else {
			full = false
		}
	}
	return full, counters[0], counters[1]
}
