package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpampServerHTTPBase(t *testing.T) {
	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "wss://acct.beta.env.middleware.io/v1/opamp", want: "https://acct.beta.env.middleware.io"},
		{in: "ws://localhost:4320/v1/opamp", want: "http://localhost:4320"},
		{in: "https://acct.middleware.io/v1/opamp?x=1#f", want: "https://acct.middleware.io"},
		{in: "ftp://acct.middleware.io/v1/opamp", wantErr: true},
		{in: "://bad", wantErr: true},
	}
	for _, c := range cases {
		got, err := opampServerHTTPBase(c.in)
		if c.wantErr {
			assert.Error(t, err, c.in)
			continue
		}
		assert.NoError(t, err, c.in)
		assert.Equal(t, c.want, got, c.in)
	}
}
