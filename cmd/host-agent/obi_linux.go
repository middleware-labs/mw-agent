package main

import (
	"github.com/middleware-labs/mw-agent/pkg/agent"
	obireceiver "go.opentelemetry.io/obi/collector"
)

// The obi (eBPF) receiver is only added here, not in pkg/agent, so agents
// that don't use it (opsai, synthetics) don't import OBI and don't need its
// generated eBPF files. It runs only when the backend config includes it
// (agent-features.cloud-network-monitoring).
func init() {
	agent.ExtraReceivers = append(agent.ExtraReceivers, obireceiver.NewFactory())
}
