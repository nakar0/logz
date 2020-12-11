package logz_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/glassonion1/logz"
	"github.com/google/go-cmp/cmp"
)

/*
Tests logz functions.
The log format is below.
{
  "severity":"INFO",
  "message":"writes info log",
  "time":"2020-12-31T23:59:59.999999999Z",
  "logging.googleapis.com/sourceLocation":{
    "file":"logz_test.go",
    "line":"46",
    "function":"github.com/glassonion1/logz_test.TestLogz.func2"
  },
  "logging.googleapis.com/trace":"projects/test/traces/00000000000000000000000000000000",
  "logging.googleapis.com/spanId":"0000000000000000",
  "logging.googleapis.com/trace_sampled":false
}
*/
func TestLogz(t *testing.T) {

	ctx := context.Background()

	now := time.Date(2020, 12, 31, 23, 59, 59, 999999999, time.UTC)
	logz.SetNow(now)
	logz.SetProjectID("test")

	// Evacuates the stdout
	orgStdout := os.Stdout
	defer func() {
		os.Stdout = orgStdout
	}()
	t.Run("Tests the Infof function", func(t *testing.T) {
		// Overrides the stdout to the buffer.
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Tests the function
		logz.Infof(ctx, "writes %s log", "info")

		w.Close()

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(r); err != nil {
			t.Fatalf("failed to read buf: %v", err)
		}

		// Gets the log from buffer.
		got := strings.TrimRight(buf.String(), "\n")

		expected := `{"severity":"INFO","message":"writes info log","time":"2020-12-31T23:59:59.999999999Z","logging.googleapis.com/sourceLocation":{"file":"logz_test.go","line":"46","function":"github.com/glassonion1/logz_test.TestLogz.func2"},"logging.googleapis.com/trace":"projects/test/traces/00000000000000000000000000000000","logging.googleapis.com/spanId":"0000000000000000","logging.googleapis.com/trace_sampled":false}`

		if diff := cmp.Diff(got, expected); diff != diff {
			// Restores the stdout
			os.Stdout = orgStdout
			t.Errorf("failed log info test: %v", diff)
		}
	})
}
