package serverfixture

import (
	"context"
	"log/slog"
	"sync"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/strotz/runner"
)

type Fixture struct {
	Server *runner.Process // Server process
}

func (f *Fixture) StartServer(ctx context.Context, waitDone *sync.WaitGroup) error {
	l, err := runfiles.Rlocation("chainsaw/link/server/server_/server")
	if err != nil {
		return err
	}
	slog.Info("Server location", "location", l)
	args := []string{}
	slog.Info("Server args:", "args")
	app, err := runner.NewProcess(ctx, l, args...)
	if err != nil {
		return err
	}
	f.Server = app
	return app.RunWithMarker(ctx, waitDone, "Server started...")
}
