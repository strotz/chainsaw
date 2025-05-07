package link

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/def"
	"google.golang.org/protobuf/proto"
)

func initTest() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
}

// Verify that pub sub is working
func TestTableCreate(t *testing.T) {
	initTest()
	ta := newTable()
	go ta.run()
	defer ta.stop()
}

func TestTableAddListener(t *testing.T) {
	initTest()
	ta := newTable()
	go ta.run()
	defer ta.stop()

	in := make(chan *def.Event)
	ta.addRecipient("test", in)
	require.Equal(t, 1, ta.length())
}

func TestTableRemoveListener(t *testing.T) {
	initTest()
	ta := newTable()
	go ta.run()
	defer ta.stop()

	in := make(chan *def.Event)
	ta.addRecipient("test", in)
	ta.remove("test")

	require.Equal(t, 0, ta.length())
}

// Publish should remove subscription
func TestTablePublishEvent(t *testing.T) {
	initTest()
	ta := newTable()
	go ta.run()
	defer ta.stop()

	// Need to buffer the response
	in := make(chan *def.Event, 1)
	ta.addRecipient("test", in)
	ta.post("test", &def.Event{})
	require.Equal(t, 0, ta.length())
	res := <-in
	expected := &def.Event{}
	require.True(t, proto.Equal(expected, res), "Expected: %v, got: %v", expected, res)
}

// Publish ignores missing recipient
func TestTablePublishNoRecipient(t *testing.T) {
	initTest()
	ta := newTable()
	go ta.run()
	defer ta.stop()

	in := make(chan *def.Event)
	ta.addRecipient("test", in)
	ta.post("test1", &def.Event{})
	require.Equal(t, 1, ta.length())
}
