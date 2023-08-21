package mwinsight

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
)

type BackendType int

type CtxKey int

const (
	TimeStampCtxKey CtxKey = 0
)

const (
	BackendTypeOpenAI = 0
	analysisChanSize  = 1000
)

// K8sInsight implements Insight interface and provides
// insights on Kubernetes errors
type K8sInsight struct {
	apiKey       string
	target       string
	k8sClient    *kubernetes.Client
	k8sNameSpace string
	backend      BackendType
	aiClient     ai.IAI
	logger       *zap.Logger
}

type K8sInsightOptionFunc func(k *K8sInsight)

// WithK8sInsightApiKey sets the unique api key for calling
// Middleware APIs.
func WithK8sInsightAPIKey(s string) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.apiKey = s
	}
}

// WithK8sInsightTarget sets target URL for sending insights
// to the Middlware backend.
func WithK8sInsightTarget(t string) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.target = t
	}
}

// WithK8sInsightK8sClient sets the Kubernetes client used by
// Middleware Insight to collect logs from the Kubernetes cluster.
func WithK8sInsightK8sClient(c *kubernetes.Client) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.k8sClient = c
	}
}

// WithK8sInsightK8sNameSpace sets the namespace for which
// Middleware Insight will analyze the issues. Leaving this empty
// will analyze logs for all namespaces.
func WithK8sInsightK8sNameSpace(n string) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.k8sNameSpace = n
	}
}

// WithK8sInsightBackend sets the backend analyzer engine. Currently
// only Open AI is supported as the backend analyzer engine.
func WithK8sInsightBackend(b BackendType) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.backend = b
		switch b {
		case BackendTypeOpenAI:
			k.aiClient = &ai.OpenAIClient{}
		}
	}
}

// WithK8sInsightLogger sets the zap logger
func WithK8sInsightLogger(l *zap.Logger) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.logger = l
	}
}

// NewK8sInsight returns new K8sInsight to be used for analyzing
// issues on Kubernetes platforms.
func NewK8sInsight(opts ...K8sInsightOptionFunc) *K8sInsight {
	var k8sInsight K8sInsight
	k8sInsight.backend = BackendTypeOpenAI
	for _, apply := range opts {
		apply(&k8sInsight)
	}

	return &k8sInsight
}

// Analyze will look for issues in the given Kubernetes clusters and
// provide insights into them for faster resolution.
func (k *K8sInsight) Analyze(ctx context.Context) (
	<-chan []byte, error) {

	// analysisChan is where results will be put for caller to process them.
	analysisChan := make(chan []byte, analysisChanSize)

	config := &analyzer.AnalysisConfiguration{
		Namespace: k.k8sNameSpace,
		NoCache:   false,
		Explain:   true,
	}

	var analysisResults *[]analyzer.Analysis = &[]analyzer.Analysis{}

	// run the analysis
	if err := analyzer.RunAnalysis(ctx, []string{}, config, k.k8sClient,
		k.aiClient, analysisResults); err != nil {
		return analysisChan, err
	}

	// concurrently process the results
	go func(analysisResults *[]analyzer.Analysis) {
		// close analysisChan so that caller can exit processing
		// it
		defer close(analysisChan)
		var innerWg sync.WaitGroup
		for _, analysis := range *analysisResults {
			// a given result might have multiple errors. Process
			// them concurrently
			for _, err := range analysis.Error {
				innerWg.Add(1)
				go func(message string, analysis analyzer.Analysis) {
					defer innerWg.Done()
					b, er := k.getPayloadToSend(ctx, message, analysis)
					if er != nil {
						k.logger.Error("error marshalling insight analysis to json", zap.Error(er))
						return
					}
					analysisChan <- b

				}(err, analysis)

			}
		}
		innerWg.Wait()
	}(analysisResults)

	return analysisChan, nil
}

// getPayloadToSend creates otlp log payload to send to
// the Middleware backend
func (k *K8sInsight) getPayloadToSend(ctx context.Context,
	message string, analysis analyzer.Analysis) ([]byte, error) {

	logs := plog.NewLogs()

	resourceLogs := logs.ResourceLogs().AppendEmpty()
	resource := resourceLogs.Resource()

	resourceAttributes := resource.Attributes()
	resourceAttributes.PutStr("mw.account_key", k.apiKey)
	resourceAttributes.PutStr("mw.resource_type", "custom")
	resourceAttributes.PutStr("service.name", analysis.Name)

	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	logAttributes := logRecord.Attributes()
	logAttributes.PutStr("component", analysis.Name)
	logAttributes.PutStr("parent", analysis.ParentObject)

	logBody := logRecord.Body()
	logBody.SetStr(message)
	logRecord.SetSeverityNumber(plog.SeverityNumberError)
	logRecord.SetSeverityText(plog.SeverityNumberError.String())

	timestamp, ok := ctx.Value(TimeStampCtxKey).(time.Time)
	if !ok {
		timestamp = time.Now()
	}

	pcommonTimeStamp := pcommon.NewTimestampFromTime(timestamp)
	logRecord.SetTimestamp(pcommonTimeStamp)
	logRecord.SetObservedTimestamp(pcommonTimeStamp)

	otlpLogReq := plogotlp.NewExportRequestFromLogs(logs)

	b, err := otlpLogReq.MarshalJSON()
	if err != nil {
		return b, err
	}
	return b, nil
}

// Send method sends a given byte slice with insight information to the Middleware backend
func (k *K8sInsight) Send(_ context.Context, data []byte) error {

	request, err := http.NewRequest("POST", k.target+"/v1/logs", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("authorization", k.apiKey)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
