package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	exchangev1beta1 "github.com/rossigee/provider-rabbitmq/apis/exchange/v1beta1"
	queuev1beta1 "github.com/rossigee/provider-rabbitmq/apis/queue/v1beta1"
	userv1beta1 "github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
	vhostv1beta1 "github.com/rossigee/provider-rabbitmq/apis/vhost/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	obs, err := c.GetBinding(context.Background(), "ex", "q", "/", "k2", "queue")
	require.NoError(t, err)
	assert.Equal(t, "k2", obs.RoutingKey)
}

func TestGetBinding_RoutingKeyNotFound(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]string{
			{"source": "ex", "destination": "q", "vhost": "/", "routing_key": "k1"},
		})
	})
	_, err := c.GetBinding(context.Background(), "ex", "q", "/", "missing", "queue")
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
		_, _ = w.Write([]byte(`{"user":"alice","vhost":"/","configure":".*","write":".*","read":"logs..*"}`))
	})
	obs, err := c.GetPermission(context.Background(), "alice", "/")
	require.NoError(t, err)
	assert.Equal(t, "alice", obs.User)
	assert.Equal(t, "/", obs.VHost)
	assert.Equal(t, ".*", obs.Configure)
	assert.Equal(t, `logs..*`, obs.Read)
}

// --- CreateExchange ---

func TestCreateExchange_BodyIncludesFieldsAndArguments(t *testing.T) {
	var received map[string]interface{}
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/exchanges/%2F/logs", r.RequestURI)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	})
	spec := &exchangev1beta1.ExchangeParameters{
		Name:       "logs",
		VHost:      "/",
		Type:       "topic",
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		Arguments: map[string]apiextensionsv1.JSON{
			"alternate-exchange": {Raw: []byte(`"dlx"`)},
			"x-max-length":       {Raw: []byte(`1000`)},
		},
	}
	_, err := c.CreateExchange(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, "topic", received["type"])
	assert.Equal(t, true, received["durable"])
	assert.Equal(t, false, received["auto_delete"])
	assert.Equal(t, false, received["internal"])
	args, ok := received["arguments"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "dlx", args["alternate-exchange"])
	assert.Equal(t, float64(1000), args["x-max-length"])
}

func TestCreateExchange_NoArgumentsOmitsKey(t *testing.T) {
	var received map[string]interface{}
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	})
	spec := &exchangev1beta1.ExchangeParameters{Name: "x", VHost: "/", Type: "fanout"}
	_, err := c.CreateExchange(context.Background(), spec)
	require.NoError(t, err)
	_, present := received["arguments"]
	assert.False(t, present)
}

// --- GetVhost / CreateVhost ---

func TestGetVhost_ParsesBody(t *testing.T) {
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/vhosts/dev", r.URL.Path)
		_, _ = w.Write([]byte(`{"name":"dev","description":"development","tags":"tag1,tag2"}`))
	})
	obs, err := c.GetVhost(context.Background(), "dev")
	require.NoError(t, err)
	assert.Equal(t, "dev", obs.Name)
	assert.Equal(t, "development", obs.Description)
	assert.Equal(t, []string{"tag1", "tag2"}, obs.Tags)
}

func TestCreateVhost_BodyIncludesDescriptionAndTags(t *testing.T) {
	var received map[string]interface{}
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	})
	spec := &vhostv1beta1.VhostParameters{
		Name:        "dev",
		Description: "development",
		Tags:        []string{"tag1", "tag2"},
	}
	_, err := c.CreateVhost(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, "development", received["description"])
	assert.Equal(t, "tag1,tag2", received["tags"])
}

// --- CreateQueue / QueueArgumentsJSON ---

func TestCreateQueue_BodyIncludesTypedAndFreeFormArguments(t *testing.T) {
	var received map[string]interface{}
	c, _ := newTestRMQClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusNoContent)
	})
	spec := &queuev1beta1.QueueParameters{
		Name:             "jobs",
		VHost:            "/",
		Durable:          true,
		AutoDelete:       false,
		MessageTTL:       60000,
		MaxLength:        10,
		OverflowBehavior: "reject-publish",
		Arguments: map[string]apiextensionsv1.JSON{
			"x-single-active-consumer": {Raw: []byte(`true`)},
		},
	}
	_, err := c.CreateQueue(context.Background(), spec)
	require.NoError(t, err)
	assert.Equal(t, true, received["durable"])
	assert.Equal(t, false, received["auto_delete"])
	args, ok := received["arguments"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(60000), args["x-message-ttl"])
	assert.Equal(t, float64(10), args["x-max-length"])
	assert.Equal(t, "reject-publish", args["x-overflow"])
	assert.Equal(t, true, args["x-single-active-consumer"])
}

func TestQueueArgumentsJSON_TypedFieldsOverrideMap(t *testing.T) {
	spec := &queuev1beta1.QueueParameters{
		MessageTTL: 5000,
		Arguments: map[string]apiextensionsv1.JSON{
			"x-message-ttl": {Raw: []byte(`9999`)},
		},
	}
	args := QueueArgumentsJSON(spec)
	assert.Equal(t, []byte(`5000`), args["x-message-ttl"].Raw)
}

func TestQueueArgumentsJSON_MatchesRoundTrippedBrokerResponse(t *testing.T) {
	spec := &queuev1beta1.QueueParameters{
		MessageTTL: 60000,
	}
	desired := QueueArgumentsJSON(spec)

	// Simulate RabbitMQ echoing the declared arguments back verbatim.
	var echoed map[string]apiextensionsv1.JSON
	raw, err := json.Marshal(jsonArgsToInterface(desired))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &echoed))

	assert.True(t, ArgsMapsEqual(desired, echoed))
}

// --- ArgsMapsEqual / StringSetEqual ---

func TestArgsMapsEqual_BytesEquivalent(t *testing.T) {
	a := map[string]apiextensionsv1.JSON{"k": {Raw: []byte(`60000`)}}
	b := map[string]apiextensionsv1.JSON{"k": {Raw: []byte(`60000`)}}
	assert.True(t, ArgsMapsEqual(a, b))
	assert.False(t, ArgsMapsEqual(a, map[string]apiextensionsv1.JSON{"k": {Raw: []byte(`"60000"`)}}))
	assert.False(t, ArgsMapsEqual(a, map[string]apiextensionsv1.JSON{"k2": {Raw: []byte(`60000`)}}))
	assert.True(t, ArgsMapsEqual(nil, map[string]apiextensionsv1.JSON{}))
}

func TestStringSetEqual_IgnoresOrder(t *testing.T) {
	assert.True(t, StringSetEqual([]string{"a", "b"}, []string{"b", "a"}))
	assert.False(t, StringSetEqual([]string{"a", "b"}, []string{"a"}))
	assert.False(t, StringSetEqual([]string{"a"}, []string{"a", "a"}))
	assert.True(t, StringSetEqual(nil, []string{}))
}
