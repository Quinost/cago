package resp2

import (
	"fmt"
	"io"
)

type RESPWriter struct {
	writer io.Writer
}

func NewRESPWriter(w io.Writer) *RESPWriter {
	return &RESPWriter{writer: w}
}

func (w *RESPWriter) WriteSimpleString(val string) error {
	_, err := fmt.Fprintf(w.writer, "+%s\r\n", val)
	return err
}

func (w *RESPWriter) WriteError(val string) error {
	_, err := fmt.Fprintf(w.writer, "-%s\r\n", val)
	return err
}

func (w *RESPWriter) WriteInteger(val int64) error {
	_, err := fmt.Fprintf(w.writer, ":%d\r\n", val)
	return err
}

func (w *RESPWriter) WriteBulkString(val string) error {
	_, err := fmt.Fprintf(w.writer, "$%d\r\n%s\r\n", len(val), val)
	return err
}

func (w *RESPWriter) WriteNull() error {
	_, err := fmt.Fprintf(w.writer, "$-1\r\n")
	return err
}

func (w *RESPWriter) WriteArray(val int) error {
	_, err := fmt.Fprintf(w.writer, "*%d\r\n", val)
	return err
}

func (w *RESPWriter) WriteNullArray() error {
	_, err := fmt.Fprintf(w.writer, "*-1\r\n")
	return err
}
