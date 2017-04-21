package table

import (
	"fmt"
	"io"
	"strings"
)

type Writer struct {
	w         io.Writer
	emptyStr  string
	bgStr     string
	borderStr string

	rows   [][]writerCell
	widths map[int]int
}

type writerCell struct {
	Value  Value
	String string
}

type hasCustomWriter interface {
	Fprintf(io.Writer, string, ...interface{}) (int, error)
}

func NewWriter(w io.Writer, emptyStr, bgStr, borderStr string) *Writer {
	return &Writer{
		w:         w,
		emptyStr:  emptyStr,
		bgStr:     bgStr,
		borderStr: borderStr,
		widths:    map[int]int{},
	}
}

func (w *Writer) Write(headers []Header, vals []Value) {
	rowsToAdd := 1
	colsWithRows := [][]writerCell{}

	visibleHeaderIndex := 0
	for i, val := range vals {
		if len(headers) > 0 && headers[i].Hidden {
			continue
		}

		var rowsInCol []writerCell

		cleanStr := strings.Replace(val.String(), "\r", "", -1)
		lines := strings.Split(cleanStr, "\n")

		if len(lines) == 1 && lines[0] == "" {
			rowsInCol = append(rowsInCol, writerCell{Value: val, String: w.emptyStr})
		} else {
			for _, line := range lines {
				rowsInCol = append(rowsInCol, writerCell{Value: val, String: line})
			}
		}

		rowsInColLen := len(rowsInCol)

		for _, cell := range rowsInCol {
			if len(cell.String) > w.widths[visibleHeaderIndex] {
				w.widths[visibleHeaderIndex] = len(cell.String)
			}
		}

		colsWithRows = append(colsWithRows, rowsInCol)

		if rowsInColLen > rowsToAdd {
			rowsToAdd = rowsInColLen
		}

		visibleHeaderIndex++
	}

	for i := 0; i < rowsToAdd; i++ {
		var row []writerCell

		for _, col := range colsWithRows {
			if i < len(col) {
				row = append(row, col[i])
			} else {
				row = append(row, writerCell{})
			}
		}

		w.rows = append(w.rows, row)
	}
}

func (w *Writer) Flush() error {
	for _, row := range w.rows {
		for colIdx, col := range row {
			if customWriter, ok := col.Value.(hasCustomWriter); ok {
				_, err := customWriter.Fprintf(w.w, "%s", col.String)
				if err != nil {
					return err
				}
			} else {
				_, err := fmt.Fprintf(w.w, "%s", col.String)
				if err != nil {
					return err
				}
			}

			paddingSize := w.widths[colIdx] - len(col.String)

			_, err := fmt.Fprintf(w.w, strings.Repeat(w.bgStr, paddingSize)+w.borderStr)
			if err != nil {
				return err
			}
		}

		_, err := fmt.Fprintln(w.w)
		if err != nil {
			return err
		}
	}

	return nil
}
