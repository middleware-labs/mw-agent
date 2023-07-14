package mwinsight

import "context"

// Insight interface provides methods to be implemented
// by different analyzers to provide insights.
type Insight interface {
	Analyze(ctx context.Context) <-chan []byte
	Send(ctx context.Context) error
}
