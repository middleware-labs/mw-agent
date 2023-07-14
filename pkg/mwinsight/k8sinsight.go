package mwinsight

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
)

type BackendType int

const (
	BackendTypeOpenAI = 0
)

type K8sInsight struct {
	apiKey       string
	target       string
	k8sClient    *kubernetes.Client
	k8sNameSpace string
	backend      BackendType
	aiClient     ai.IAI
}

type K8sInsightOptionFunc func(k *K8sInsight)

func WithK8sInsightApiKey(s string) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.apiKey = s
	}
}

func WithK8sInsightTarget(t string) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.target = t
	}
}

func WithK8sInsightK8sClient(c *kubernetes.Client) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.k8sClient = c
	}
}

func WithK8sInsightK8sNameSpace(n string) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.k8sNameSpace = n
	}
}

func WithK8sInsightBackend(b BackendType) K8sInsightOptionFunc {
	return func(k *K8sInsight) {
		k.backend = b
		switch b {
		case BackendTypeOpenAI:
			k.aiClient = &ai.OpenAIClient{}
		}
	}
}

func NewK8sInsight(opts ...K8sInsightOptionFunc) *K8sInsight {
	var analyzer K8sInsight
	analyzer.backend = BackendTypeOpenAI
	for _, apply := range opts {
		apply(&analyzer)
	}

	return &analyzer
}

func (k *K8sInsight) Analyze(ctx context.Context) (
	<-chan []byte, error) {
	analysisChan := make(chan []byte)

	config := &analyzer.AnalysisConfiguration{
		Namespace: k.k8sNameSpace,
		NoCache:   false,
		Explain:   true,
	}

	var analysisResults *[]analyzer.Analysis = &[]analyzer.Analysis{}
	if err := analyzer.RunAnalysis(ctx, []string{}, config, k.k8sClient,
		k.aiClient, analysisResults); err != nil {
		return analysisChan, err
	}

	go func(analysisResults *[]analyzer.Analysis) {
		// close analysisChan so that caller can exit processing
		// it
		defer close(analysisChan)
		var innerWg sync.WaitGroup
		for _, analysis := range *analysisResults {
			for _, err := range analysis.Error {
				innerWg.Add(1)
				go func(message string, analysis analyzer.Analysis) {
					defer innerWg.Done()

					analysisChan <- k.getPayloadToSend(ctx, message, analysis)

				}(err, analysis)

			}

		}
		innerWg.Wait()
	}(analysisResults)

	return analysisChan, nil
}

func (k *K8sInsight) getPayloadToSend(ctx context.Context,
	message string, analysis analyzer.Analysis) []byte {

	// TODO create a resource struct instead of a string json
	return []byte(`{
			"resource_logs": [
			  {
				"resource": {
				  "attributes": [
					{
					  "key": "mw.account_key",
					  "value": {
						"string_value": "` + k.apiKey + `"
					  }
					},
					{
					  "key": "mw.resource_type",
					  "value": {
						"string_value": "custom"
					  }
					},
					{
					  "key": "service.name",
					  "value": {
						"string_value": "` + analysis.Name + `"
					  }
					}
				  ]
				},
				"scope_logs": [
					{
					  "log_records": [
						  {
							  "attributes": [
								{
								  "key": "component",
								  "value": {
									"string_value": "` + analysis.Name + `"
								  }
								},
								{
									"key": "parent",
									"value": {
									  "string_value": "` + analysis.ParentObject + `"
									}
								  }
							  ],
							  "body": {
								  "string_value": "` + message + `"
								 
							  },
							  "severity_number": 17,
							  "severity_text": "ERROR",
							  "time_unix_nano": ` + strconv.FormatInt(time.Now().UnixNano(), 10) + `,
							  "observed_time_unix_nano": ` + strconv.FormatInt(time.Now().UnixNano(), 10) + `
						  }
					  ]   
					}
				 ]
			  }
			]
		  } 
		  `)
}

func (k *K8sInsight) Send(ctx context.Context, data []byte) error {

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
