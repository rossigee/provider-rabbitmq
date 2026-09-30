package clients

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetUser_DecodesArrayTags is a regression test for the fault that stopped
// every User reconcile.
//
// GetUser decoded the response into a struct with `Tags string`, but RabbitMQ
// returns a JSON array. That produced, on every reconcile:
//
//	failed to unmarshal user: json: cannot unmarshal array into Go struct field .tags of type string
//
// Both payload shapes are taken from a live instance at
// https://queues.bankrut.lan/api/users/.
func TestGetUser_DecodesArrayTags(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "array form, as returned by a live instance",
			body: `{"name":"vm-events","tags":[],"limits":{}}`,
			want: []string{},
		},
		{
			name: "populated array",
			body: `{"name":"admin","tags":["administrator"]}`,
			want: []string{"administrator"},
		},
		{
			name: "multi element array",
			body: `{"name":"ops","tags":["monitoring","policymaker"]}`,
			want: []string{"monitoring", "policymaker"},
		},
		{
			name: "comma separated string form is still tolerated",
			body: `{"name":"ops","tags":"monitoring,policymaker"}`,
			want: []string{"monitoring", "policymaker"},
		},
		{
			name: "tags absent",
			body: `{"name":"ops"}`,
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/users/ops", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			})

			obs, err := c.GetUser(t.Context(), "ops")
			require.NoError(t, err, "GetUser must not error on a well-formed response")
			assert.Equal(t, tc.want, obs.Tags)
		})
	}
}

// TestGetVhost_DecodesArrayTags covers the same fault on the vhost path. The
// released v0.5.2 image did not unmarshal the vhost response at all, so this
// only started failing once real vhost observation landed; master has the same
// string-typed field and would have regressed on the next release.
func TestGetVhost_DecodesArrayTags(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "array form, as returned by a live instance",
			body: `{"name":"accounts","description":"","tags":[],"tracer_port":null}`,
			want: []string{},
		},
		{
			name: "populated array",
			body: `{"name":"ops","description":"d","tags":["monitoring"],"tracer_port":5672}`,
			want: []string{"monitoring"},
		},
		{
			name: "comma separated string form is still tolerated",
			body: `{"name":"ops","description":"d","tags":"monitoring,policymaker","tracer_port":5672}`,
			want: []string{"monitoring", "policymaker"},
		},
		{
			name: "tags absent",
			body: `{"name":"ops","description":"d","tracer_port":5672}`,
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			})

			obs, err := c.GetVhost(t.Context(), "ops")
			require.NoError(t, err, "GetVhost must not error on a well-formed response")
			assert.Equal(t, tc.want, obs.Tags)
		})
	}
}
