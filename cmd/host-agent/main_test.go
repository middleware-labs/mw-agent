package main

import (
	"errors"
	"fmt"
	"testing"

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
