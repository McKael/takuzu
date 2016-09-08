package takuzu

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var verbosity int
var schrodLvl uint

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

	// TODO: validate chars ([.01OI])

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

// DumpString writes the content of the board as a stream
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

func (b Takuzu) guessPos(l, c int) int {
	if b.Board[l][c].Defined {
		return b.Board[l][c].Value
	}

	bx := b.Clone()
	bx.Set(l, c, 0)
	bx.FillLineColumn(l, c)
	if bx.CheckLine(l) != nil || bx.CheckColumn(c) != nil {
		return 1
	}
	Copy(&b, &bx)
	bx.Set(l, c, 1)
	bx.FillLineColumn(l, c)
	if bx.CheckLine(l) != nil || bx.CheckColumn(c) != nil {
		return 0
	}

	return -1 // dunno
}

// TrivialHint returns the coordinates and the value of the first cell that
// can be guessed using trivial methods.
// It returns {-1, -1, -1} if none can be found.
func (b Takuzu) TrivialHint() (line, col, value int) {
	for line = 0; line < b.Size; line++ {
		for col = 0; col < b.Size; col++ {
			if b.Board[line][col].Defined {
				continue
			}
			if value = b.guessPos(line, col); value != -1 {
				return
			}
		}
	}
	value, line, col = -1, -1, -1
	return
}

// trySolveTrivialPass does 1 pass over the takuzu board and tries to find
// values using simple guesses.
func (b Takuzu) trySolveTrivialPass() (changed bool) {
	for line := 0; line < b.Size; line++ {
		for col := 0; col < b.Size; col++ {
			if b.Board[line][col].Defined {
				continue
			}
			if guess := b.guessPos(line, col); guess != -1 {
				b.Set(line, col, guess)
				if verbosity > 3 {
					log.Printf("Trivial: Setting [%d,%d] to %d", line, col, guess)
				}
				changed = true // Ideally remember l,c
			}
		}
	}
	return changed
}

// TrySolveTrivial tries to solve the takuzu using a loop over simple methods
// It returns true if all cells are defined, and an error if the grid breaks the rules.
func (b Takuzu) TrySolveTrivial() (bool, error) {
	for {
		changed := b.trySolveTrivialPass()
		if verbosity > 3 {
			var status string
			if changed {
				status = "ongoing"
			} else {
				status = "stuck"
			}
			log.Println("Trivial resolution -", status)
		}
		if !changed {
			break
		}

		if verbosity > 3 {
			b.DumpBoard()
			fmt.Println()
		}
	}
	full, err := b.Validate()
	if err != nil {
		return full, errors.Wrap(err, "the takuzu looks wrong")
	}
	return full, nil
}

// TrySolveRecurse tries to solve the takuzu recursively, using trivial
// method first and using guesses if it fails.
func (b Takuzu) TrySolveRecurse(allSolutions *[]Takuzu, timeout time.Duration) (*Takuzu, error) {

	var solutionsMux sync.Mutex
	var singleSolution *Takuzu
	var solutionMap map[string]*Takuzu

	var globalSearch bool
	// globalSearch doesn't need to use a mutex and is more convenient
	// to use than allSolutions.
	if allSolutions != nil {
		globalSearch = true
		solutionMap = make(map[string]*Takuzu)
	}

	startTime := time.Now()

	var recurseSolve func(level int, t Takuzu, errStatus chan<- error) error

	recurseSolve = func(level int, t Takuzu, errStatus chan<- error) error {

		reportStatus := func(failure error) {
			// Report status to the caller's channel
			if errStatus != nil {
				errStatus <- failure
			}
		}

		// In Schröndinger mode we check concurrently both values for a cell
		var schrodinger bool
		concurrentRoutines := 1
		if level < int(schrodLvl) {
			schrodinger = true
			concurrentRoutines = 2
		}

		var status [2]chan error
		status[0] = make(chan error)
		status[1] = make(chan error)

		for {
			// Try simple resolution first
			full, err := t.TrySolveTrivial()
			if err != nil {
				reportStatus(err)
				return err
			}

			if full { // We're done
				if verbosity > 1 {
					log.Printf("{%d} The takuzu is correct and complete.", level)
				}
				solutionsMux.Lock()
				singleSolution = &t
				if globalSearch {
					solutionMap[t.ToString()] = &t
				}
				solutionsMux.Unlock()

				reportStatus(nil)
				return nil
			}

			if verbosity > 2 {
				log.Printf("{%d} Trivial resolution did not complete.", level)
			}

			// Trivial method is stuck, let's use recursion

			changed := false

			// Looking for first empty cell
			var line, col int
		firstClear:
			for line = 0; line < t.Size; line++ {
				for col = 0; col < t.Size; col++ {
					if !t.Board[line][col].Defined {
						break firstClear
					}
				}
			}

			if line == t.Size || col == t.Size {
				break
			}

			if verbosity > 2 {
				log.Printf("{%d} GUESS - Trying values for [%d,%d]", level, line, col)
			}

			var val int
			err = nil
			errCount := 0

			for testval := 0; testval < 2; testval++ {
				if !globalSearch && t.Board[line][col].Defined {
					// No need to "guess" here anymore
					break
				}

				// Launch goroutines for cell values of 0 and/or 1
				for testCase := 0; testCase < 2; testCase++ {
					if schrodinger || testval == testCase {
						tx := t.Clone()
						tx.Set(line, col, testCase)
						go recurseSolve(level+1, tx, status[testCase])
					}
				}

				// Let's collect the goroutines' results
				for i := 0; i < concurrentRoutines; i++ {
					if schrodinger && verbosity > 1 { // XXX
						log.Printf("{%d} Schrodinger waiting for result #%d for cell [%d,%d]", level, i, line, col)
					}
					select {
					case e := <-status[0]:
						err = e
						val = 0
					case e := <-status[1]:
						err = e
						val = 1
					}

					if schrodinger && verbosity > 1 { // XXX
						log.Printf("{%d} Schrodinger result #%d/2 for cell [%d,%d]=%d - err=%v", level, i+1, line, col, val, err)
					}

					if err == nil {
						if !globalSearch {
							reportStatus(nil)
							if i+1 < concurrentRoutines {
								// Schröndinger mode and we still have one status to fetch
								<-status[1-val]
							}
							return nil
						}
						continue
					}
					if timeout > 0 && level > 2 && time.Since(startTime) > timeout {
						if errors.Cause(err).Error() != "timeout" {
							if verbosity > 0 {
								log.Printf("{%d} Timeout, giving up", level)
							}
							err := errors.New("timeout")
							reportStatus(err)
							if i+1 < concurrentRoutines {
								// Schröndinger mode and we still have one status to fetch
								<-status[1-val]
							}
							// XXX actually can't close the channel and leave, can I?
							return err
						}
					}

					// err != nil: we can set a value --  unless this was a timeout
					if errors.Cause(err).Error() == "timeout" {
						if verbosity > 1 {
							log.Printf("{%d} Timeout propagation", level)
						}
						reportStatus(err)
						if i+1 < concurrentRoutines {
							// Schröndinger mode and we still have one status to fetch
							<-status[1-val]
						}
						// XXX actually can't close the channel and leave, can I?
						return err
					}

					errCount++
					if verbosity > 2 {
						log.Printf("{%d} Bad outcome (%v)", level, err)
						log.Printf("{%d} GUESS was wrong - Setting [%d,%d] to %d",
							level, line, col, 1-val)
					}

					t.Set(line, col, 1-val)
					changed = true

				} // concurrentRoutines

				if (changed && !globalSearch) || schrodinger {
					// Let's loop again with the new board
					break
				}
			}

			if verbosity > 2 {
				log.Printf("{%d} End of cycle.\n\n", level)
			}

			if errCount == 2 {
				// Both values failed
				err := errors.New("dead end")
				reportStatus(err)
				return err
			}

			// If we cannot fill more cells (!changed) or if we've made a global search with
			// both values, the search is complete.
			if schrodinger || globalSearch || !changed {
				break
			}

			if verbosity > 2 {
				t.DumpBoard()
				fmt.Println()
			}
		}

		// Try to force garbage collection
		runtime.GC()

		full, err := t.Validate()
		if err != nil {
			if verbosity > 1 {
				log.Println("The takuzu looks wrong - ", err)
			}
			err := errors.Wrap(err, "the takuzu looks wrong")
			reportStatus(err)
			return err
		}
		if full {
			if verbosity > 1 {
				log.Println("The takuzu is correct and complete")
			}
			solutionsMux.Lock()
			singleSolution = &t
			if globalSearch {
				solutionMap[t.ToString()] = &t
			}
			solutionsMux.Unlock()
		}

		reportStatus(nil)
		return nil
	}

	status := make(chan error)
	go recurseSolve(0, b, status)

	err := <-status // Wait for it...

	firstSol := singleSolution
	if globalSearch {
		for _, tp := range solutionMap {
			*allSolutions = append(*allSolutions, *tp)
		}
	}

	if err != nil {
		return firstSol, err
	}

	if globalSearch && len(*allSolutions) > 0 {
		firstSol = &(*allSolutions)[0]
	}
	return firstSol, nil
}

// SetSchrodingerLevel initializes the "Schrödinger" level (0 means disabled)
// It must be called before any board generation or reduction.
func SetSchrodingerLevel(level uint) {
	schrodLvl = level
}

// SetVerbosityLevel initializes the verbosity level
func SetVerbosityLevel(level int) {
	verbosity = level
}
