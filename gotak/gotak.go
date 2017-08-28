// Copyright (C) 2016 Mikael Berthe <mikael@lilotux.net>. All rights reserved.
// Use of this source code is governed by the MIT license,
// which can be found in the LICENSE file.

// gotak is a CLI wrapper for the takuzu package.

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/pflag"

	"hg.lilotux.net/golang/mikael/takuzu"
)

var verbosity int

func newTakuzuGameBoard(size int, simple bool, jobs int, buildBoardTimeout, reduceBoardTimeout time.Duration, minRatio, maxRatio int) *takuzu.Takuzu {
	results := make(chan *takuzu.Takuzu)

	newTak := func(i int) {
		takuzu, err := takuzu.NewRandomTakuzu(size, simple, fmt.Sprintf("%v", i),
			buildBoardTimeout, reduceBoardTimeout, minRatio, maxRatio)

		if err == nil && takuzu != nil {
			results <- takuzu
			if verbosity > 0 && jobs > 1 {
				log.Printf("Worker #%d done.", i)
			}
		} else {
			results <- nil
		}
	}

	if jobs == 0 {
		return nil
	}
	for i := 0; i < jobs; i++ {
		go newTak(i)
	}
	tak := <-results
	return tak
}

func main() {
	vbl := pflag.Uint("vl", 0, "Verbosity Level")
	simple := pflag.Bool("simple", false, "Only look for trivial solutions")
	out := pflag.Bool("out", false, "Send solution string to output")
	board := pflag.String("board", "", "Load board string")
	schrodLvl := pflag.Uint("x-sl", 0, "[Advanced] Schrödinger level")
	resolveTimeout := pflag.Duration("x-timeout", 0, "[Advanced] Resolution timeout")
	buildBoardTimeout := pflag.Duration("x-build-timeout", 5*time.Minute, "[Advanced] Build timeout per resolution")
	reduceBoardTimeout := pflag.Duration("x-reduce-timeout", 20*time.Minute, "[Advanced] Reduction timeout")
	buildMinRatio := pflag.Uint("x-new-min-ratio", 55, "[Advanced] Build empty cell ratio (40-60)")
	buildMaxRatio := pflag.Uint("x-new-max-ratio", 62, "[Advanced] Build empty cell ratio (50-99)")
	all := pflag.Bool("all", false, "Look for all possible solutions")
	reduce := pflag.Bool("reduce", false, "Try to reduce the number of digits")
	buildNewSize := pflag.Uint("new", 0, "Build a new takuzu board (with given size)")
	pdfFileName := pflag.String("to-pdf", "", "PDF output file name")
	workers := pflag.Uint("workers", 1, "Number of parallel workers (use with --new)")

	pflag.Parse()

	verbosity = int(*vbl)
	takuzu.SetVerbosityLevel(verbosity)
	takuzu.SetSchrodingerLevel(*schrodLvl)

	var tak *takuzu.Takuzu

	if *board != "" {
		var err error
		tak, err = takuzu.NewFromString(*board)
		if tak == nil || err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			tak = nil
		}
	}

	if *buildNewSize > 0 {
		if verbosity > 1 {
			log.Printf("buildBoardTimeout:   %v", *buildBoardTimeout)
			log.Printf("reduceBoardTimeout:  %v", *reduceBoardTimeout)
			log.Printf("Free cell min ratio: %v", *buildMinRatio)
			log.Printf("Free cell max ratio: %v", *buildMaxRatio)
		}
		tak = newTakuzuGameBoard(int(*buildNewSize), *simple,
			int(*workers),
			*buildBoardTimeout, *reduceBoardTimeout,
			int(*buildMinRatio), int(*buildMaxRatio))
	}

	if tak == nil {
		fmt.Fprintln(os.Stderr, "Could not create takuzu board.")
		os.Exit(255)
	}

	tak.DumpBoard()
	fmt.Println()

	if *pdfFileName != "" {
		if err := tak2pdf(tak, *pdfFileName); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if *out {
			tak.DumpString()
		}
		os.Exit(0)
	}

	if *buildNewSize > 0 {
		if *out {
			tak.DumpString()
		}
		os.Exit(0)
	}

	if *reduce {
		if verbosity > 1 {
			log.Printf("buildBoardTimeout:   %v", *buildBoardTimeout)
			log.Printf("reduceBoardTimeout:  %v", *reduceBoardTimeout)
		}
		var err error
		if tak, err = tak.ReduceBoard(*simple, "0", *buildBoardTimeout, *reduceBoardTimeout); err != nil {
			log.Println(err)
			os.Exit(1)
		}

		tak.DumpBoard()
		fmt.Println()

		if *out {
			tak.DumpString()
		}

		os.Exit(0)
	}

	if *simple {
		full, err := tak.TrySolveTrivial()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if !full {
			tak.DumpBoard()
			fmt.Println()
			if *out {
				tak.DumpString()
			}
			log.Println("The takuzu could not be completed using trivial methods.")
			os.Exit(2)
		}

		log.Println("The takuzu is correct and complete.")
		tak.DumpBoard()
		fmt.Println()

		if *out {
			tak.DumpString()
		}
		os.Exit(0)
	}

	var allSol *[]takuzu.Takuzu
	if *all {
		allSol = &[]takuzu.Takuzu{}
	}
	res, err := tak.TrySolveRecurse(allSol, *resolveTimeout)
	if err != nil && verbosity > 1 {
		// The last trivial resolution failed
		log.Println("Trivial resolution failed:", err)
	}

	// Ignoring res & err if a full search was requested
	if *all {
		log.Println(len(*allSol), "solution(s) found.")
		if len(*allSol) > 0 {
			for _, s := range *allSol {
				if *out {
					s.DumpString()
				} else {
					s.DumpBoard()
					fmt.Println()
				}
			}
			if len(*allSol) > 1 {
				os.Exit(3)
			}
			os.Exit(0)
		}
		fmt.Println("No solution found.")
		os.Exit(2)
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if res != nil {
		res.DumpBoard()
		fmt.Println()

		if *out {
			res.DumpString()
		}
		os.Exit(0)
	}

	fmt.Println("No solution found.")
	os.Exit(2)
}
