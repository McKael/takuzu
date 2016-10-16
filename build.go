// Copyright (C) 2016 Mikael Berthe <mikael@lilotux.net>. All rights reserved.
// Use of this source code is governed by the MIT license,
// which can be found in the LICENSE file.

package takuzu

// This file contains the functions and methods used to build a new takuzu
// puzzle.

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type buildTakuzuOptions struct {
	size                                  int
	minRatio, maxRatio                    int
	simple                                bool
	buildBoardTimeout, reduceBoardTimeout time.Duration
}

// ReduceBoard randomly removes as many numbers as possible from the
// takuzu board and returns a pointer to the new board.
// The initial takuzu might be modified.
func (tak Takuzu) ReduceBoard(trivial bool, wid string, buildBoardTimeout, reduceBoardTimeout time.Duration) (*Takuzu, error) {

	size := tak.Size

	// First check if the board is correct
	if verbosity > 0 {
		log.Printf("[%v]ReduceBoard: Checking for all grid solutions...", wid)
	}

	allSol := &[]Takuzu{}
	_, err := tak.Clone().TrySolveRecurse(allSol, buildBoardTimeout)
	ns := len(*allSol)
	if err != nil && errors.Cause(err).Error() == "timeout" {
		if verbosity > 0 {
			log.Printf("[%v]ReduceBoard: There was a timeout (%d resolution(s) found).", wid, ns)
		}
		if ns == 0 {
			return nil, err
		}
		//if ns < 10 { return nil, err }
		if verbosity > 0 {
			log.Printf("[%v]ReduceBoard: Going on anyway...", wid)
		}
	}

	if verbosity > 0 {
		log.Printf("[%v]ReduceBoard: %d solution(s) found.", wid, ns)
	}

	if ns == 0 {
		return nil, err
	} else if ns > 1 {
		tak = (*allSol)[rand.Intn(ns)]
		if verbosity > 0 {
			log.Printf("[%v]ReduceBoard: Warning: there are %d solutions.", wid, ns)
			log.Printf("[%v]ReduceBoard: Picking one randomly.", wid)

			if verbosity > 1 {
				tak.DumpBoard()
				fmt.Println()
			}
		}
		allSol = nil
	} else {
		// 1 and only 1 solution
		if verbosity > 1 {
			tak.DumpBoard()
			fmt.Println()
		}
	}

	if verbosity > 0 {
		log.Printf("[%v]ReduceBoard: Grid reduction...", wid)
	}
	fields := make([]*Cell, size*size)
	n := 0
	for l := range tak.Board {
		for c := range tak.Board[l] {
			if tak.Board[l][c].Defined {
				fields[n] = &tak.Board[l][c]
				n++
			}
		}
	}

	nDigits := 0
	initialDigits := n
	ratio := 0
	if verbosity > 0 {
		log.Printf("[%v]ReduceBoard: %d%%", wid, ratio)
	}

	for ; n > 0; n-- {
		var rollback bool
		i := rand.Intn(n)
		fields[i].Defined = false
		if trivial {
			full, err := tak.Clone().TrySolveTrivial()
			if err != nil || !full {
				rollback = true
			}
		} else {
			allSol = &[]Takuzu{}
			_, err := tak.Clone().TrySolveRecurse(allSol, reduceBoardTimeout)
			if err != nil || len(*allSol) != 1 {
				rollback = true
			}
		}

		if rollback {
			if verbosity > 1 {
				log.Printf("[%v]ReduceBoard: Backing out", wid)
			}
			fields[i].Defined = true // Back out!
			nDigits++
		}
		fields = append(fields[:i], fields[i+1:]...)

		if verbosity > 0 {
			nr := (initialDigits - n) * 100 / initialDigits
			if nr > ratio {
				ratio = nr
				log.Printf("[%v]ReduceBoard: %d%%", wid, ratio)
			}
		}
	}

	if verbosity > 0 {
		log.Printf("[%v]ReduceBoard: I have left %d digits.", wid, nDigits)
	}

	return &tak, nil
}

// newRandomTakuzu creates a new Takuzu board with a given size
// It is intended to be called by NewRandomTakuzu only.
func newRandomTakuzu(wid string, buildOpts buildTakuzuOptions) (*Takuzu, error) {
	size := buildOpts.size
	easy := buildOpts.simple
	buildBoardTimeout := buildOpts.buildBoardTimeout
	reduceBoardTimeout := buildOpts.reduceBoardTimeout
	minRatio := buildOpts.minRatio
	maxRatio := buildOpts.maxRatio

	tak := New(size)
	n := size * size
	fields := make([]*Cell, n)

	i := 0
	for l := range tak.Board {
		for c := range tak.Board[l] {
			fields[i] = &tak.Board[l][c]
			i++
		}
	}

	if verbosity > 0 {
		log.Printf("[%v]NewRandomTakuzu: Filling new board (%dx%[2]d)...", wid, size)
	}

	nop := 0

	// #1. Loop until the ratio of empty cells is less than minRatio% (e.g. 55%)

	for n > size*size*minRatio/100 {
		i := rand.Intn(n)
		fields[i].Defined = true
		fields[i].Value = rand.Intn(2)

		var err error

		if _, err = tak.Validate(); err != nil {
			if verbosity > 1 {
				log.Printf("[%v]NewRandomTakuzu: Could not set cell value to %v", wid, fields[i].Value)
			}
		} else if _, err = tak.Clone().TrySolveTrivial(); err != nil {
			if verbosity > 1 {
				log.Printf("[%v]NewRandomTakuzu: Trivial checks: Could not set cell value to %v", wid, fields[i].Value)
			}
		}

		if err == nil {
			fields = append(fields[:i], fields[i+1:]...)
			n--
			continue
		}

		// If any of the above checks fails, we roll back
		fields[i].Defined = false
		fields[i].Value = 0 // Let's reset but it is useless

		// Safety check to avoid deadlock on bad boards
		nop++
		if nop > 2*size*size {
			log.Printf("[%v]NewRandomTakuzu: Could not fill up board!", wid)
			// Givin up on this board
			return nil, errors.New("could not fill up board") // Try again
		}

	}

	var ptak *Takuzu
	var removed int

	// #2. Try to solve the current board; try to remove some cells if it fails

	// Initial empty cells count
	iecc := n

	for {
		// Current count of empty (i.e. undefined) cells
		ec := iecc + removed
		ecpc := ec * 100 / (size * size)
		if verbosity > 0 {
			log.Printf("[%v]NewRandomTakuzu: Empty cells: %d (%d%%)", wid, ec, ecpc)
		}
		if ecpc > maxRatio {
			if verbosity > 0 {
				log.Printf("[%v]NewRandomTakuzu: Too many empty cells (%d); giving up on this board", wid, ec)
			}
			break
		}
		var err error
		ptak, err = tak.ReduceBoard(easy, wid, buildBoardTimeout, reduceBoardTimeout)
		if err != nil && errors.Cause(err).Error() == "timeout" {
			break
		}
		if err == nil && ptak != nil {
			break
		}
		if verbosity > 0 {
			log.Printf("[%v]NewRandomTakuzu: Could not use this grid", wid)
		}
		inc := size * size / 150
		if inc == 0 {
			inc = 1
		}
		tak.removeRandomCell(inc)
		removed += inc
		if verbosity > 1 {
			log.Printf("[%v]NewRandomTakuzu: Removed %d numbers", wid, removed)
			if verbosity > 1 {
				tak.DumpBoard()
			}
		}
	}

	if ptak == nil {
		if verbosity > 0 {
			log.Printf("[%v]NewRandomTakuzu: Couldn't use this board, restarting from scratch...", wid)
		}
		return nil, errors.New("could not use current board") // Try again
	}

	return ptak, nil
}

// NewRandomTakuzu creates a new Takuzu board with a given size
func NewRandomTakuzu(size int, simple bool, wid string, buildBoardTimeout, reduceBoardTimeout time.Duration, minRatio, maxRatio int) (*Takuzu, error) {
	if size%2 != 0 {
		return nil, errors.New("board size should be an even value")
	}

	if size < 4 {
		return nil, errors.New("board size is too small")
	}

	// minRatio : percentage (1-100) of empty cells when creating a new board
	// If the board is wrong the cells will be removed until we reach maxRatio

	if minRatio < 40 {
		minRatio = 40
	}
	if minRatio > maxRatio {
		return nil, errors.New("minRatio/maxRatio incorrect")
	}

	if maxRatio > 99 {
		maxRatio = 99
	}

	buildOptions := buildTakuzuOptions{
		size:               size,
		minRatio:           minRatio,
		maxRatio:           maxRatio,
		simple:             simple,
		buildBoardTimeout:  buildBoardTimeout,
		reduceBoardTimeout: reduceBoardTimeout,
	}

	var takP *Takuzu

	for {
		var err error
		takP, err = newRandomTakuzu(wid, buildOptions)
		if err == nil {
			break
		}
	}

	return takP, nil
}

func (tak Takuzu) removeRandomCell(number int) {
	size := tak.Size
	fields := make([]*Cell, size*size)
	n := 0
	for l := range tak.Board {
		for c := range tak.Board[l] {
			if tak.Board[l][c].Defined {
				fields[n] = &tak.Board[l][c]
				n++
			}
		}
	}
	for i := 0; i < number; i++ {
		if n == 0 {
			return
		}
		fields[rand.Intn(n)].Defined = false
		fields = append(fields[:i], fields[i+1:]...)
		n--
	}
}
