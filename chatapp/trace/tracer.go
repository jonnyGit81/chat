package trace

import (
	"fmt"
	"io"
)


// Tracer is the interface that describes an object capable of
// tracing events throughout code.
type Tracer interface {
	Trace(...interface{})
}

func New(w io.Writer) Tracer  {
	return &tracer{out: w}
}

// New creates a new Tracer that will write the output to
// the specified io.Writer.
// implement tracer interface.
// New above we cannot return interface, we must return the concrete type
// we are not exported but it return at the Tracer interface New. this is fine
type tracer struct {
	out io.Writer
}

// tracer is a Tracer that writes to an
// io.Writer.
func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprint(t.out, a...)
	fmt.Fprintln(t.out)
}


// For user not used any tracer we set as nilTracer which is to do nothing,
// then we initialize this to factory method in the room
type nilTracer struct {}

func (n *nilTracer) Trace(...interface{}) {}

func Off() Tracer  {
	return &nilTracer{}
}

