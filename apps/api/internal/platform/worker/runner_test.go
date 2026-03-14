package worker

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestHealthcheck_Output(t *testing.T) {
	oldStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = writePipe
	Healthcheck("scraper")
	_ = writePipe.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(readPipe)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	expected := []byte("worker=scraper status=ok")
	if !bytes.Contains(output, expected) {
		t.Fatalf("expected output to contain %q, got %q", expected, output)
	}
}

func TestRun_ReturnsNilOnContextCancel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, logger, "scraper", 50*time.Millisecond)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error on graceful stop, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker run did not stop after context cancel")
	}
}
