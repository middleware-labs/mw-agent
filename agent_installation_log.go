package main

import (
	// "context"

	"os"
	// "go.opentelemetry.io/collector/pdata/plog"
	// "go.opentelemetry.io/collector/pdata/plog/plogotlp"
	// "time"
)

func agent_installation_log() {

	if err := os.MkdirAll("/tmp/mwagent", os.ModePerm); err != nil {
		// log.Fatal(err)
	}

	f, _ := os.OpenFile("/tmp/mwagent/install_success.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)

	defer f.Close()

	f.WriteString("Middleware Agent Installed Successfully !\n")

	// ld := plog.NewLogs()
	// rl := ld.ResourceLogs().AppendEmpty()
	// sl := rl.ScopeLogs().AppendEmpty()
	// lr := sl.LogRecords().AppendEmpty()

	// resourceAttrs := rl.Resource().Attributes()
	// resourceAttrs.EnsureCapacity(1)

	// hostname, _ := os.Hostname()

	// fmt.Println("hostname", hostname)
	// fmt.Println("hostname222", os.Getenv("MW_API_KEY"))

	// resourceAttrs.PutStr("host.id", hostname)
	// resourceAttrs.PutStr("mw.account_key", os.Getenv("MW_API_KEY"))
	// resourceAttrs.PutStr("agent.installation.time", time.Now().String())

	// // A static message to show that the middleware agent is running
	// lr.Body().SetStr("Middleware Agent installed successfully")

	// // Log added as type "info"
	// lr.SetSeverityNumber(9)

	// req := plogotlp.NewExportRequestFromLogs(ld)

	// middlewareLogExporterFactory := NewFactory()
	// exporters = []exporter.Factory{
	// 	middlewareLogExporterFactory,
	// }

	// factories.Exporters, err = exporter.MakeFactoryMap(exporters...)
	// if err != nil {
	// 	return otelcol.Factories{}, err
	// }

	// logExporter, _ := newMiddlewareLogExporter()
	// _, err := logExporter.Export(context.Background(), req)

}
