package def

func MakeEvent(p isEvent_Payload) *Event {
	return &Event{
		Payload: p,
	}
}

func MakeEnvelope(id string, p isEvent_Payload) *Envelope {
	return &Envelope{
		CallId: &CallId{
			Id: id,
		},
		Event: MakeEvent(p),
	}
}

func WrapEvent(id string, e *Event) *Envelope {
	return &Envelope{
		CallId: &CallId{Id: id},
		Event:  e,
	}
}
