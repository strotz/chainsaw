package link

import (
	"log/slog"

	"github.com/strotz/chainsaw/link/def"
)

// TODO: redo API namings. It is not pub/sub, because pub effectively
// removes the subscriber.

type tableItem struct {
	id  string
	sub chan<- *def.Event
}

type idAndValue struct {
	id  string
	val *def.Event
}

type table struct {
	addCh    chan *tableItem
	removeCh chan string
	notifyCh chan *idAndValue
	stopCh   chan struct{}
	queryCh  chan struct{}
	respCh   chan int
	subs     map[string]*tableItem
}

func newTable() *table {
	t := &table{
		addCh:    make(chan *tableItem),
		removeCh: make(chan string),
		notifyCh: make(chan *idAndValue),
		stopCh:   make(chan struct{}),
		queryCh:  make(chan struct{}),
		respCh:   make(chan int),
		subs:     make(map[string]*tableItem),
	}
	return t
}

// TODO: should it be keyed by payload type?

func (t *table) addRecipient(id string, sub chan<- *def.Event) {
	t.addCh <- &tableItem{
		id:  id,
		sub: sub,
	}
}

func (t *table) remove(id string) {
	t.removeCh <- id
}

func (t *table) post(id string, payload *def.Event) {
	slog.Debug("posting event", "id", id, "payload", payload)
	t.notifyCh <- &idAndValue{
		id:  id,
		val: payload,
	}
}

func (t *table) length() int {
	t.queryCh <- struct{}{}
	return <-t.respCh
}

func (t *table) run() {
	for {
		select {
		case <-t.stopCh:
			return
		case item := <-t.addCh:
			slog.Debug("Add element", "id", item.id)
			t.subs[item.id] = item
		case id := <-t.removeCh:
			slog.Debug("Remove element", "id", id)
			delete(t.subs, id)
		case v := <-t.notifyCh:
			slog.Debug("Notify element", "id", v.id)
			r, found := t.subs[v.id]
			if found {
				slog.Debug("Notified", "id", r.id, "val", v.val)
				r.sub <- v.val
				delete(t.subs, v.id)
			} else {
				slog.Debug("Not found element", "id", v.id)
			}
		case <-t.queryCh:
			slog.Debug("Query length")
			t.respCh <- len(t.subs)
		}
	}
}

func (t *table) stop() {
	close(t.stopCh)
}
