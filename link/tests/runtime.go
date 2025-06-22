package tests

import (
	"context"
	"flag"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/stretchr/testify/require"
)

type Runtime struct {
	// Global context for the test
	Ctx      context.Context
	Cancel   context.CancelFunc
	WaitDone sync.WaitGroup
	name     string
}

func Setup(t *testing.T) *Runtime {
	flag.Parse()
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))

	tmpDir := filepath.Join(bazel.TestTmpDir(), t.Name())
	require.NoError(t, os.MkdirAll(tmpDir, fs.FileMode(0755)))

	r := &Runtime{
		name: t.Name(),
	}
	r.Ctx, r.Cancel = context.WithCancel(context.Background())
	slog.Debug("---------TEST:", "name", r.name)
	return r
}

func (r *Runtime) WithTimeout(timeout time.Duration) *Runtime {
	r.Ctx, r.Cancel = context.WithTimeout(r.Ctx, timeout)
	return r
}

func (r *Runtime) Close() {
	r.Cancel()
	r.WaitDone.Wait()
	slog.Debug("---------STOP:", "name", r.name)
}

// WaitFor periodically executes p until it returns true.
func (r *Runtime) WaitFor(p func() bool) error {
	for {
		if p() {
			return nil
		}
		select {
		case <-r.Ctx.Done():
			return r.Ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
