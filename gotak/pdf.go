// Copyright (C) 2016 Mikael Berthe <mikael@lilotux.net>. All rights reserved.
// Use of this source code is governed by the MIT license,
// which can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/pkg/errors"

	"github.com/McKael/takuzu"
)

func tak2pdf(takuzu *takuzu.Takuzu, pdfFileName string) error {

	if pdfFileName == "" {
		return errors.New("no PDF file name")
	}

	size := takuzu.Size

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 14)

	basicTable := func() {

		for ln, l := range takuzu.Board {
			for cn, cell := range l {
				border := "" // empty, "1", "L", "T", "R" and "B"
				if ln == 0 {
					border += "T"
				}
				if cn == 0 {
					border += "L"
				}
				if ln+1 == size {
					border += "B"
				}
				if cn+1 == size {
					border += "R"
				}
				align := "CM" // horiz=Center vert=Middle
				if cell.Defined {
					pdf.CellFormat(8, 8, fmt.Sprint(cell.Value), border, 0, align, false, 0, "")
				} else {
					pdf.CellFormat(8, 8, ".", border, 0, align, false, 0, "")
				}
			}
			pdf.Ln(-1)
		}
	}

	pdf.AddPage()
	basicTable()
	if err := pdf.OutputFileAndClose(pdfFileName); err != nil {
		return err
	}

	return nil
}
