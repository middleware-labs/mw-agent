package main

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeCollector struct {
	calls []string
}

func (f *fakeCollector) StartCollector() error {
	f.calls = append(f.calls, "start")
	return nil
}

func (f *fakeCollector) StopCollector(err error) {
	f.calls = append(f.calls, "stop")
}

func TestControlCollector(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "failed restart-status check keeps the collector running",
			err:  fmt.Errorf("%w: restart api returned non-200 status: 502", agent.ErrRestartStatusCheckFailure),
			want: nil,
		},
		{
			name: "config fetch failure keeps the collector running",
			err:  fmt.Errorf("%w: boom", agent.ErrConfigFetchFailure),
			want: nil,
		},
		{
			name: "restart",
			err:  agent.ErrRestartAgent,
			want: []string{"stop", "start"},
		},
		{
			name: "nil starts",
			err:  nil,
			want: []string{"start"},
		},
		{
			name: "other error stops",
			err:  errors.New("something else"),
			want: []string{"stop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &program{logger: zap.NewNop(), errCh: make(chan error)}
			fake := &fakeCollector{}
			done := make(chan struct{})
			go func() {
				p.controlCollector(fake)
				close(done)
			}()

			p.errCh <- tt.err
			close(p.errCh)
			<-done

			assert.Equal(t, tt.want, fake.calls)
		})
	}
}

// Stop must not close errCh while a sender may still use it: it closes stopCh,
// waits for the senders, then closes errCh. Run with -race.
func TestProgramStopClosesErrChAfterSenders(t *testing.T) {
	p := &program{
		logger:    zap.NewNop(),
		programWG: &sync.WaitGroup{},
		errCh:     make(chan error),
		stopCh:    make(chan struct{}),
	}
	p.programWG.Add(1)
	go func() { // the reader, like run
		for range p.errCh {
		}
		p.programWG.Done()
	}()
	for i := 0; i < 4; i++ { // senders that keep reporting until stopped
		p.programWG.Add(1)
		p.senders.Add(1)
		go func() {
			defer p.programWG.Done()
			defer p.senders.Done()
			for {
				select {
				case <-p.stopCh:
					return
				case p.errCh <- nil:
				}
			}
		}()
	}

	stopped := make(chan struct{})
	go func() { _ = p.Stop(nil); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return")
	}
}
