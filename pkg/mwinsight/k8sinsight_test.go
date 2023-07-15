package mwinsight

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyze(t *testing.T) {

	tests := []struct {
		name           string
		clientset      *fake.Clientset
		apiKey         string
		target         string
		namespace      string
		timestamp      time.Time
		compareCount   int
		expectedResult string
		err            error
	}{
		{
			name: "valid test",
			clientset: fake.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "example",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{
						{
							Type:    v1.PodScheduled,
							Reason:  "Unschedulable",
							Message: "0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate.",
						},
					},
				},
			}),
			apiKey:    "12345",
			target:    "https://devnull.mw.lc",
			namespace: "default",
			timestamp: func() time.Time {
				const layout = "Jan 2, 2006 at 3:04pm (PST)"

				// Calling Parse() method with its parameters
				tm, _ := time.Parse(layout, "Jul 14, 2023 at 11:00pm (PST)")
				return tm
			}(),
			compareCount:   1,
			expectedResult: `{"resourceLogs":[{"resource":{"attributes":[{"key":"mw.account_key","value":{"stringValue":"1234"}},{"key":"mw.resource_type","value":{"stringValue":"custom"}},{"key":"service.name","value":{"stringValue":"default/example"}}]},"scopeLogs":[{"scope":{},"logRecords":[{"timeUnixNano":"1689375600000000000","observedTimeUnixNano":"1689375600000000000","severityNumber":17,"severityText":"Error","body":{"stringValue":"0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate."},"attributes":[{"key":"component","value":{"stringValue":"default/example"}},{"key":"parent","value":{"stringValue":"example"}}],"traceId":"","spanId":""}]}]}]}`,
			err:            nil,
		},

		{
			name: "no results due to namespace difference",
			clientset: fake.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "example",
					Namespace:   "default-different-ns",
					Annotations: map[string]string{},
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{
						{
							Type:    v1.PodScheduled,
							Reason:  "Unschedulable",
							Message: "0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate.",
						},
					},
				},
			}),
			apiKey:    "12345",
			target:    "https://devnull.mw.lc",
			namespace: "default",
			timestamp: func() time.Time {
				const layout = "Jan 2, 2006 at 3:04pm (PST)"

				// Calling Parse() method with its parameters
				tm, _ := time.Parse(layout, "Jul 14, 2023 at 11:00pm (PST)")
				return tm
			}(),
			compareCount:   0,
			expectedResult: `{"resourceLogs":[{"resource":{"attributes":[{"key":"mw.account_key","value":{"stringValue":"1234"}},{"key":"mw.resource_type","value":{"stringValue":"custom"}},{"key":"service.name","value":{"stringValue":"default/example"}}]},"scopeLogs":[{"scope":{},"logRecords":[{"timeUnixNano":"1689375600000000000","observedTimeUnixNano":"1689375600000000000","severityNumber":17,"severityText":"Error","body":{"stringValue":"0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate."},"attributes":[{"key":"component","value":{"stringValue":"default/example"}},{"key":"parent","value":{"stringValue":"example"}}],"traceId":"","spanId":""}]}]}]}`,
			err:            nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			k8sClient := &kubernetes.Client{
				Client: test.clientset,
			}

			insight := NewK8sInsight(
				WithK8sInsightK8sClient(k8sClient),
				WithK8sInsightBackend(BackendTypeOpenAI),
				WithK8sInsightApiKey("1234"),
				WithK8sInsightTarget("https://test.middleware.io"),
				WithK8sInsightK8sNameSpace(test.namespace),
			)

			ctx := context.WithValue(context.Background(), TimeStampCtxKey,
				test.timestamp)
			analysisChan, err := insight.Analyze(ctx)
			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, test.err, err)
			}

			if err != nil {
				return
			}

			compareCount := 0
			for analysis := range analysisChan {
				compareCount++
				assert.Equal(t, test.expectedResult, string(analysis))
			}
			assert.Equal(t, test.compareCount, compareCount)
		})
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		httpMethod  string
		contentType string
		serverURL   string
		server      *httptest.Server
		err         error
	}{
		{
			name:        "valid test",
			apiKey:      "12345",
			httpMethod:  "POST",
			contentType: "application/json",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Assert that the request contains the expected headers and body
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "12345", r.Header.Get("authorization"))
				assert.Equal(t, "POST", r.Method)

				// Read the request body
				body, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)

				// Assert the request body contains the expected data
				expectedData := []byte("some data")
				assert.Equal(t, expectedData, body)

				// Respond with a success status code
				w.WriteHeader(http.StatusOK)
			})),
			err: nil,
		},
		{
			name:        "invalid URL",
			apiKey:      "12345",
			httpMethod:  "POST",
			contentType: "application/json",
			serverURL:   "nonscheme://www.mw.lc",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Assert that the request contains the expected headers and body
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "12345", r.Header.Get("authorization"))
				assert.Equal(t, "POST", r.Method)

				// Read the request body
				body, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)

				// Assert the request body contains the expected data
				expectedData := []byte("some data")
				assert.Equal(t, expectedData, body)

				// Respond with a success status code
				w.WriteHeader(http.StatusOK)
			})),
			err: errors.New("Post \"nonscheme://www.mw.lc/v1/logs\": unsupported protocol scheme \"nonscheme\""),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			logger := zap.NewNop()
			serverURL := test.serverURL
			if serverURL == "" {
				serverURL = test.server.URL
			}
			insight := NewK8sInsight(
				WithK8sInsightBackend(BackendTypeOpenAI),
				WithK8sInsightApiKey(test.apiKey),
				WithK8sInsightTarget(serverURL),
				WithK8sInsightLogger(logger),
			)

			ctx := context.Background()
			err := insight.Send(ctx, []byte("some data"))
			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualValues(t, test.err.Error(), err.Error())
			}
		})
	}

}
