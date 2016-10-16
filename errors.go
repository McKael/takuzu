package takuzu

import "fmt"

// This file contains the takuzu validation error type.

const (
	ErrorNil = iota
	ErrorDuplicate
	ErrorTooManyValues
	ErrorTooManyAdjacentValues
)

type validationError struct {
	ErrorType    int
	LineNumber   *int
	ColumnNumber *int
	CellValue    *int
}

func (e validationError) Error() string {
	var axis string
	var n int

	// Currently we don't have validation errors with both
	// line and column so we can get the axis:
	if e.LineNumber != nil {
		axis = "line"
		n = *e.LineNumber
	} else if e.ColumnNumber != nil {
		axis = "column"
		n = *e.ColumnNumber
	}

	switch e.ErrorType {
	case ErrorNil:
		return ""
	case ErrorDuplicate:
		if axis == "" {
			return "internal validation error"
		}
		return fmt.Sprintf("duplicate %ss (%d)", axis, n)
	case ErrorTooManyValues:
		if axis == "" || e.CellValue == nil {
			return "internal validation error"
		}
		var numberStr string
		if *e.CellValue == 0 {
			numberStr = "zeroes"
		} else if *e.CellValue == 1 {
			numberStr = "ones"
		} else {
			return "internal validation error"
		}
		return fmt.Sprintf("%s %d: too many %s", axis, n, numberStr)
	case ErrorTooManyAdjacentValues:
		if axis == "" || e.CellValue == nil {
			return "internal validation error"
		}
		return fmt.Sprintf("%s %d: 3+ same values %d", axis, n, *e.CellValue)
	}
	return "internal validation error"
}
