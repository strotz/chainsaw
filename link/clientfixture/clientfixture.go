package clientfixture

import (
	"context"
	"errors"
	"log"
	"log/slog"

	"github.com/strotz/chainsaw/link"
	"github.com/strotz/chainsaw/link/tests"
)

type Fixture struct {
	Client *link.Client
}

func (f *Fixture) RunConnected(r *tests.Runtime) error {
	if f.Client != nil {
		return errors.New("client fixture already used")
	}
	c, err := link.NewClient()
	if err != nil {
		return err
	}
	f.Client = c

	r.WaitDone.Add(1)
	go func() {
		defer r.WaitDone.Done()
		err := c.Start(r.Ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalln("Client exists with unexpected error", "err", err)
		}
	}()

	// Wait for the client to connect. It is necessary, to avoid error from c.Start()
	err = r.WaitFor(func() bool {
		return c.Connected.Load()
	})
	if err != nil {
		_ = c.Close()
		f.Client = nil
		return err
	}

	slog.Debug("Connected")
	return nil
}

func (f *Fixture) Close() {
	if f.Client == nil {
		slog.Debug("Client is nil")
		return
	}
	err := f.Client.Close()
	slog.Debug("Closed", "err", err)
}
