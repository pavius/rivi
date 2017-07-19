package sd

import (
	"io"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/instance"
)

func TestDefaultEndpointer(t *testing.T) {
	var (
		ca = make(closer)
		cb = make(closer)
		c  = map[string]io.Closer{"a": ca, "b": cb}
		f  = func(instance string) (endpoint.Endpoint, io.Closer, error) {
			return endpoint.Nop, c[instance], nil
		}
		instancer = &mockInstancer{
			cache: instance.NewCache(),
		}
	)
	// set initial state
	instancer.Update(sd.Event{Instances: []string{"a", "b"}})

	endpointer := sd.NewEndpointer(instancer, f, log.NewNopLogger(), sd.InvalidateOnError(time.Minute))
	if endpoints, err := endpointer.Endpoints(); err != nil {
		t.Errorf("unepected error %v", err)
	} else if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	instancer.Update(sd.Event{Instances: []string{}})
	select {
	case <-ca:
		t.Logf("endpoint a closed, good")
	case <-time.After(time.Millisecond):
		t.Errorf("didn't close the deleted instance in time")
	}
	select {
	case <-cb:
		t.Logf("endpoint b closed, good")
	case <-time.After(time.Millisecond):
		t.Errorf("didn't close the deleted instance in time")
	}
	if endpoints, err := endpointer.Endpoints(); err != nil {
		t.Errorf("unepected error %v", err)
	} else if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	endpointer.Close()
	instancer.Update(sd.Event{Instances: []string{"a"}})
	// TODO verify that on Close the endpointer fully disconnects from the instancer.
	// Unfortunately, because we use instance.Cache, this test cannot be in the sd package,
	// and therefore does not have access to the endpointer's private members.
}

type mockInstancer struct {
	cache *instance.Cache
}

func (m *mockInstancer) Update(event sd.Event) {
	m.cache.Update(event)
}

func (m *mockInstancer) Register(ch chan<- sd.Event) {
	m.cache.Register(ch)
}

func (m *mockInstancer) Deregister(ch chan<- sd.Event) {
	m.cache.Deregister(ch)
}

type closer chan struct{}

func (c closer) Close() error { close(c); return nil }