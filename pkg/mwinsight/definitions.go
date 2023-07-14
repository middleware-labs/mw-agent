package mwinsight

import "context"

type Insight interface {
	Analyze(ctx context.Context) <-chan []byte
	Send(ctx context.Context) error
}
