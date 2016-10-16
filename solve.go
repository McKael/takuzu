// Copyright (C) 2016 Mikael Berthe <mikael@lilotux.net>. All rights reserved.
// Use of this source code is governed by the MIT license,
// which can be found in the LICENSE file.

package takuzu

// This file contains the methods used to solve a takuzu puzzle.

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var verbosity int
var schrodLvl uint

// SetVerbosityLevel initializes the verbosity level of the resolution
// routines.
func SetVerbosityLevel(level int) {
	verbosity = level
}

// SetSchrodingerLevel initializes the "Schrödinger" level (0 means disabled)
// It must be called before any board generation or reduction.
func SetSchrodingerLevel(level uint) {
	schrodLvl = level
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
