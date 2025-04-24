package hello

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/serverfixture"
	"github.com/strotz/chainsaw/link/tests"
)

func TestRunHello(t *testing.T) {
	r := tests.Setup(t)
	defer r.Close()

	s := serverfixture.Fixture{}
	require.NoError(t, s.StartServer(r.Ctx, &r.WaitDone))

	// TODO: create client
	// TODO: send request and receive response
}
