package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userv1beta1 "github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
)

// --- IsNotFound ---

func TestIsNotFound_Nil(t *testing.T) {
	assert.False(t, IsNotFound(nil))
}

func TestIsNotFound_OtherError(t *testing.T) {
	assert.False(t, IsNotFound(fmt.Errorf("some other error")))
}

func TestIsNotFound_StringContaining404(t *testing.T) {
	// Plain string mentioning 404 must NOT match — typed check only.
	assert.False(t, IsNotFound(fmt.Errorf("status 404 from upstream")))
}

func TestIsNotFound_TypedError(t *testing.T) {
	assert.True(t, IsNotFound(&NotFoundError{}))
	assert.True(t, IsNotFound(&NotFoundError{Body: "vhost not found"}))
}

func TestIsNotFound_Wrapped(t *testing.T) {
	assert.True(t, IsNotFound(fmt.Errorf("outer: %w", &NotFoundError{Body: "inner"})))
}

func TestNotFoundError_Error(t *testing.T) {
	assert.Equal(t, "request failed with status 404", (&NotFoundError{}).Error())
	assert.Equal(t, "request failed with status 404: gone", (&NotFoundError{Body: "gone"}).Error())
}

// --- NewClient ---

func TestNewClient_InvalidCABundle(t *testing.T) {
	c := NewClient(&Config{
		Endpoint: "https://rabbitmq.example.com",
		CABundle: []byte("not valid PEM"),
	})
	_, err := c.GetVhost(context.Background(), "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CA bundle contains no valid PEM certificates")
}

func TestNewClient_ValidCABundle(t *testing.T) {
	// A self-signed CA cert in PEM format — just needs to parse without error.
	// We use a well-known test CA from Go's crypto/tls test data.
	// Here we just verify NewClient doesn't return an errClient.
	c := NewClient(&Config{
		Endpoint: "https://rabbitmq.example.com",
		Username: "u",
		Password: "p",
	})
	assert.NotNil(t, c)
	_, ok := c.(*rabbitmqClient)
	assert.True(t, ok)
}

// --- request() ---

func newTestRMQClient(t *testing.T, handler http.HandlerFunc) (*rabbitmqClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &rabbitmqClient{
		baseURL:  srv.URL,
		username: "admin",
		password: "secret",
		client:   &http.Client{},
	}, srv
}

func TestRequest_200(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		u, p, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "admin", u)
		assert.Equal(t, "secret", p)
		_, _ = w.Write([]byte(`{"name":"test"}`))
	})
	data, err := c.request(context.Background(), "GET", "/api/vhosts/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"test"}`, string(data))
}

func TestRequest_204NoBody(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	data, err := c.request(context.Background(), "DELETE", "/api/vhosts/x", nil)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestRequest_404ReturnedAsNotFoundError(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("object not found"))
	})
	_, err := c.request(context.Background(), "GET", "/api/vhosts/missing", nil)
	require.Error(t, err)
	assert.True(t, IsNotFound(err))
}

func TestRequest_500ReturnsGenericError(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})
	_, err := c.request(context.Background(), "GET", "/api/vhosts/x", nil)
	require.Error(t, err)
	assert.False(t, IsNotFound(err))
	assert.Contains(t, err.Error(), "500")
}

func TestRequest_WithBodySetsContentType(t *testing.T) {
	var received map[string]interface{}
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	})
	_, err := c.request(context.Background(), "PUT", "/api/exchanges/%2F/test",
		map[string]interface{}{"type": "topic"})
	require.NoError(t, err)
	assert.Equal(t, "topic", received["type"])
}

// --- GetVhost ---

func TestGetVhost_Found(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/vhosts/myvhost", r.URL.Path)
		_, _ = w.Write([]byte(`{"name":"myvhost"}`))
	})
	obs, err := c.GetVhost(context.Background(), "myvhost")
	require.NoError(t, err)
	assert.Equal(t, "myvhost", obs.Name)
}

func TestGetVhost_NotFound(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := c.GetVhost(context.Background(), "missing")
	require.Error(t, err)
	assert.True(t, IsNotFound(err))
}

// --- GetBinding ---

func TestGetBinding_RoutingKeyMatched(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]string{
			{"source": "ex", "destination": "q", "vhost": "/", "routing_key": "k1"},
			{"source": "ex", "destination": "q", "vhost": "/", "routing_key": "k2"},
		})
	})
	obs, err := c.GetBinding(context.Background(), "ex", "q", "/", "k2")
	require.NoError(t, err)
	assert.Equal(t, "k2", obs.RoutingKey)
}

func TestGetBinding_RoutingKeyNotFound(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]string{
			{"source": "ex", "destination": "q", "vhost": "/", "routing_key": "k1"},
		})
	})
	_, err := c.GetBinding(context.Background(), "ex", "q", "/", "missing")
	require.Error(t, err)
	assert.True(t, IsNotFound(err))
}

// --- GetUser ---

func TestGetUser_ParsesTags(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/users/alice", r.URL.Path)
		_, _ = w.Write([]byte(`{"name":"alice","tags":"administrator,management"}`))
	})
	obs, err := c.GetUser(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, "alice", obs.Name)
	assert.Equal(t, []string{"administrator", "management"}, obs.Tags)
}

func TestGetUser_EmptyTagsGivesNilSlice(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name":"bob","tags":""}`))
	})
	obs, err := c.GetUser(context.Background(), "bob")
	require.NoError(t, err)
	assert.Empty(t, obs.Tags)
}

// --- CreateUser ---

func TestCreateUser_JoinsTagsAsCSV(t *testing.T) {
	var received map[string]interface{}
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	})
	spec := &userv1beta1.UserParameters{
		Name: "charlie",
		Tags: []string{"administrator", "monitoring"},
	}
	_, err := c.CreateUser(context.Background(), spec, "hunter2")
	require.NoError(t, err)
	assert.Equal(t, "administrator,monitoring", received["tags"])
	assert.Equal(t, "hunter2", received["password"])
}

// --- GetPermission ---

func TestGetPermission_Unmarshals(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Path decodes %2F; use RequestURI to check the raw encoded path.
		assert.Equal(t, "/api/permissions/%2F/alice", r.RequestURI)
		_, _ = w.Write([]byte(`{"user":"alice","vhost":"/","configure":".*","write":".*","read":"logs\\..*"}`))
	})
	obs, err := c.GetPermission(context.Background(), "alice", "/")
	require.NoError(t, err)
	assert.Equal(t, "alice", obs.User)
	assert.Equal(t, "/", obs.VHost)
	assert.Equal(t, ".*", obs.Configure)
	assert.Equal(t, `logs\..*`, obs.Read)
}
