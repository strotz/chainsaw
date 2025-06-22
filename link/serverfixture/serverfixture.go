package serverfixture

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/strotz/runner"
)

type Fixture struct {
	Server *runner.Process // Server process
}

// StartServer starts the server process and exits. It is expected that the server will run in the background. serverStopped
// is used to signal that the server is stopped, and it is safe to exit the test.
func (f *Fixture) StartServer(ctx context.Context, serverStopped *sync.WaitGroup) error {
	l, err := runfiles.Rlocation("chainsaw/link/server/server_/server")
	if err != nil {
		return err
	}
	slog.Info("Server location", "location", l)
	args := []string{}
	slog.Info("Server:", "args", args)
	app, err := runner.NewProcess(ctx, l, args...)
	if err != nil {
		return err
	}
	f.Server = app
	return app.RunWithMarker(ctx, serverStopped, "Server started...")
}

// SoftStop sends a SIGINT signal to the process, which is a soft stop.
func (f *Fixture) SoftStop(ctx context.Context) error {
	if f.Server == nil {
		return errors.New("server fixture is nil")
	}
	if err := f.Server.SendSignal(os.Interrupt); err != nil {
		return err
	}
	return f.Server.StdErrScanner().WaitForKeyword(ctx, "Server stopped")
}
