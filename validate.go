package takuzu

// This file contains the takuzu validation functions and methods.

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
					v := c.Value
					return full, validationError{
						ErrorType: ErrorTooManyAdjacentValues,
						CellValue: &v,
					}
				}
			} else {
				prevCellCount = 1
			}

		}
		prevCell = c
	}
	if counters[0] > size/2 {
		v := 0
		return full, validationError{
			ErrorType: ErrorTooManyValues,
			CellValue: &v,
		}
	}
	if counters[1] > size/2 {
		v := 1
		return full, validationError{
			ErrorType: ErrorTooManyValues,
			CellValue: &v,
		}
	}
	return full, nil
}

// CheckRangeCounts returns true if all cells of the provided range are defined,
// as well as the number of 0s and the number of 1s in the range.
func CheckRangeCounts(cells []Cell) (full bool, n0, n1 int) {
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
			err := err.(validationError)
			err.LineNumber = &i
			return false, err
		}
		if full {
			hv := computeVal(d)
			if lineVals[hv] {
				err := validationError{
					ErrorType:  ErrorDuplicate,
					LineNumber: &i,
				}
				return false, err
			}
			lineVals[hv] = true
		} else {
			finished = false
		}

		// Let's check column i
		d = b.GetColumn(i)
		full, err = checkRange(d)
		if err != nil {
			err := err.(validationError)
			err.ColumnNumber = &i
			return false, err
		}
		if full {
			hv := computeVal(d)
			if colVals[hv] {
				err := validationError{
					ErrorType:    ErrorDuplicate,
					ColumnNumber: &i,
				}
				return false, err
			}
			colVals[hv] = true
		} else {
			finished = false
		}
	}
	return finished, nil
}
