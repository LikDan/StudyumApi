package utils

import (
	"github.com/xuri/excelize/v2"
	"strings"
	"unicode/utf8"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NextColumn(previous string) string {
	if previous == "" {
		return "A"
	}

	i := strings.Index(alphabet, previous[len(previous)-1:])
	if i == len(alphabet)-1 {
		return NextColumn(previous[:len(previous)-1]) + "A"
	}

	return previous[:len(previous)-1] + alphabet[i+1:i+2]
}

func AutoSizeColumns(f *excelize.File, sheetName string) error {
	cols, err := f.GetCols(sheetName)
	if err != nil {
		return err
	}
	for idx, col := range cols {
		largestWidth := 0
		for _, rowCell := range col {
			cellWidth := utf8.RuneCountInString(rowCell) + 2 // + 2 for margin
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}
		name, err := excelize.ColumnNumberToName(idx + 1)
		if err != nil {
			return err
		}
		if err = f.SetColWidth(sheetName, name, name, float64(largestWidth)); err != nil {
			return err
		}
	}

	return nil
}
