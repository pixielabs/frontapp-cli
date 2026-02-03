package output

import (
	"bytes"
	"testing"
)

func TestPlainTableWriterUsesTabs(t *testing.T) {
	var buf bytes.Buffer

	tbl := NewTableWriter(&buf, true)
	tbl.AddRow("A", "B", "C")

	if err := tbl.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	if got := buf.String(); got != "A\tB\tC\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}
