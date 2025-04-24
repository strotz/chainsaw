package tests

import (
	"context"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/stretchr/testify/require"
)

type Runtime struct {
	// Global context for the test
	Ctx      context.Context
	Cancel   context.CancelFunc
	WaitDone sync.WaitGroup
}

func Setup(t *testing.T) *Runtime {
	flag.Parse()

	tmpDir := filepath.Join(bazel.TestTmpDir(), t.Name())
	require.NoError(t, os.MkdirAll(tmpDir, fs.FileMode(0755)))

	r := &Runtime{}
	r.Ctx, r.Cancel = context.WithCancel(context.Background())
	return r
}

func (r *Runtime) Close() {
	r.Cancel()
	r.WaitDone.Wait()
}
