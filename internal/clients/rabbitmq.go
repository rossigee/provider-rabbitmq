/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	pv1beta1 "github.com/rossigee/provider-rabbitmq/apis/v1beta1"
	bindingv1beta1 "github.com/rossigee/provider-rabbitmq/apis/binding/v1beta1"
	exchangev1beta1 "github.com/rossigee/provider-rabbitmq/apis/exchange/v1beta1"
	permissionv1beta1 "github.com/rossigee/provider-rabbitmq/apis/permission/v1beta1"
	queuev1beta1 "github.com/rossigee/provider-rabbitmq/apis/queue/v1beta1"
	userv1beta1 "github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
	vhostv1beta1 "github.com/rossigee/provider-rabbitmq/apis/vhost/v1beta1"
)

type Config struct {
	Endpoint string
	Username string
	Password string
	// CABundle is a PEM-encoded CA certificate pool for verifying the server
	// certificate. When nil the system CA pool is used.
	CABundle []byte
}

type Client interface {
	GetVhost(ctx context.Context, name string) (*vhostv1beta1.VhostObservation, error)
	CreateVhost(ctx context.Context, spec *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error)
	DeleteVhost(ctx context.Context, name string) error

	GetExchange(ctx context.Context, name, vhost string) (*exchangev1beta1.ExchangeObservation, error)
	CreateExchange(ctx context.Context, spec *exchangev1beta1.ExchangeParameters) (*exchangev1beta1.ExchangeObservation, error)
	DeleteExchange(ctx context.Context, name, vhost string) error

	GetQueue(ctx context.Context, name, vhost string) (*queuev1beta1.QueueObservation, error)
	CreateQueue(ctx context.Context, spec *queuev1beta1.QueueParameters) (*queuev1beta1.QueueObservation, error)
	DeleteQueue(ctx context.Context, name, vhost string) error

	GetBinding(ctx context.Context, source, destination, vhost, routingKey string) (*bindingv1beta1.BindingObservation, error)
	CreateBinding(ctx context.Context, spec *bindingv1beta1.BindingParameters) (*bindingv1beta1.BindingObservation, error)
	DeleteBinding(ctx context.Context, source, destination, vhost, routingKey string) error

	GetUser(ctx context.Context, name string) (*userv1beta1.UserObservation, error)
	CreateUser(ctx context.Context, spec *userv1beta1.UserParameters, password string) (*userv1beta1.UserObservation, error)
	DeleteUser(ctx context.Context, name string) error

	GetPermission(ctx context.Context, user, vhost string) (*permissionv1beta1.PermissionObservation, error)
	SetPermission(ctx context.Context, spec *permissionv1beta1.PermissionParameters) (*permissionv1beta1.PermissionObservation, error)
	DeletePermission(ctx context.Context, user, vhost string) error
}

type rabbitmqClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewClient(config *Config) Client {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if len(config.CABundle) > 0 {
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(config.CABundle); !ok {
			// Return a client that will fail on first request rather than
			// silently using the system pool when a custom CA was requested.
			return &errClient{err: errors.New("CA bundle contains no valid PEM certificates")}
		}
		tlsCfg.RootCAs = pool
	}
	return &rabbitmqClient{
		baseURL:  strings.TrimSuffix(config.Endpoint, "/"),
		username: config.Username,
		password: config.Password,
		client: &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		},
	}
}

// errClient is a Client that always returns a fixed error, used when
// configuration is invalid (e.g. a malformed CA bundle).
type errClient struct{ err error }

func (e *errClient) GetVhost(_ context.Context, _ string) (*vhostv1beta1.VhostObservation, error) {
	return nil, e.err
}
func (e *errClient) CreateVhost(_ context.Context, _ *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error) {
	return nil, e.err
}
func (e *errClient) DeleteVhost(_ context.Context, _ string) error { return e.err }
func (e *errClient) GetExchange(_ context.Context, _, _ string) (*exchangev1beta1.ExchangeObservation, error) {
	return nil, e.err
}
func (e *errClient) CreateExchange(_ context.Context, _ *exchangev1beta1.ExchangeParameters) (*exchangev1beta1.ExchangeObservation, error) {
	return nil, e.err
}
func (e *errClient) DeleteExchange(_ context.Context, _, _ string) error { return e.err }
func (e *errClient) GetQueue(_ context.Context, _, _ string) (*queuev1beta1.QueueObservation, error) {
	return nil, e.err
}
func (e *errClient) CreateQueue(_ context.Context, _ *queuev1beta1.QueueParameters) (*queuev1beta1.QueueObservation, error) {
	return nil, e.err
}
func (e *errClient) DeleteQueue(_ context.Context, _, _ string) error { return e.err }
func (e *errClient) GetBinding(_ context.Context, _, _, _, _ string) (*bindingv1beta1.BindingObservation, error) {
	return nil, e.err
}
func (e *errClient) CreateBinding(_ context.Context, _ *bindingv1beta1.BindingParameters) (*bindingv1beta1.BindingObservation, error) {
	return nil, e.err
}
func (e *errClient) DeleteBinding(_ context.Context, _, _, _, _ string) error { return e.err }
func (e *errClient) GetUser(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
	return nil, e.err
}
func (e *errClient) CreateUser(_ context.Context, _ *userv1beta1.UserParameters, _ string) (*userv1beta1.UserObservation, error) {
	return nil, e.err
}
func (e *errClient) DeleteUser(_ context.Context, _ string) error { return e.err }
func (e *errClient) GetPermission(_ context.Context, _, _ string) (*permissionv1beta1.PermissionObservation, error) {
	return nil, e.err
}
func (e *errClient) SetPermission(_ context.Context, _ *permissionv1beta1.PermissionParameters) (*permissionv1beta1.PermissionObservation, error) {
	return nil, e.err
}
func (e *errClient) DeletePermission(_ context.Context, _, _ string) error { return e.err }

func GetConfig(ctx context.Context, kube client.Client, mg resource.Managed) (*Config, error) {
	pc := &pv1beta1.ProviderConfig{}

	pcName := "default"
	type typedPCReferencer interface {
		GetProviderConfigReference() *xpv2.ProviderConfigReference
	}
	if r, ok := mg.(typedPCReferencer); ok {
		if ref := r.GetProviderConfigReference(); ref != nil && ref.Name != "" {
			pcName = ref.Name
		}
	}

	if err := kube.Get(ctx, types.NamespacedName{
		Name:      pcName,
		Namespace: mg.GetNamespace(),
	}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get ProviderConfig")
	}

	data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, kube, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials")
	}

	var secretData map[string]string
	if err := json.Unmarshal(data, &secretData); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal credentials")
	}

	if pc.Spec.Endpoint == "" {
		return nil, errors.New("ProviderConfig endpoint is required")
	}

	cfg := &Config{
		Endpoint: pc.Spec.Endpoint,
		Username: secretData["username"],
		Password: secretData["password"],
	}

	if pc.Spec.TLS != nil && pc.Spec.TLS.CABundleSecretRef != nil {
		caData, err := resource.CommonCredentialExtractor(
			ctx,
			xpv2.CredentialsSourceSecret,
			kube,
			xpv2.CommonCredentialSelectors{SecretRef: pc.Spec.TLS.CABundleSecretRef},
		)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get TLS CA bundle")
		}
		cfg.CABundle = caData
	}

	return cfg, nil
}

// NotFoundError is returned when a RabbitMQ API call returns HTTP 404.
type NotFoundError struct {
	Body string
}

func (e *NotFoundError) Error() string {
	if e.Body != "" {
		return "request failed with status 404: " + e.Body
	}
	return "request failed with status 404"
}

func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

func (c *rabbitmqClient) request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse URL")
	}

	var reqReader io.Reader = http.NoBody
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
		reqReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqReader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.SetBasicAuth(c.username, c.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 404 {
			return nil, &NotFoundError{Body: strings.TrimSpace(string(errBody))}
		}
		if len(errBody) > 0 {
			return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, errBody)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if resp.StatusCode == 204 {
		return nil, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	return respBody, nil
}

func (c *rabbitmqClient) GetVhost(ctx context.Context, name string) (*vhostv1beta1.VhostObservation, error) {
	path := fmt.Sprintf("/api/vhosts/%s", url.PathEscape(name))
	_, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return &vhostv1beta1.VhostObservation{Name: name}, nil
}

func (c *rabbitmqClient) CreateVhost(ctx context.Context, spec *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error) {
	path := fmt.Sprintf("/api/vhosts/%s", url.PathEscape(spec.Name))
	_, err := c.request(ctx, "PUT", path, nil)
	if err != nil {
		return nil, err
	}
	return &vhostv1beta1.VhostObservation{Name: spec.Name}, nil
}

func (c *rabbitmqClient) DeleteVhost(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/vhosts/%s", url.PathEscape(name))
	_, err := c.request(ctx, "DELETE", path, nil)
	return err
}

func (c *rabbitmqClient) GetExchange(ctx context.Context, name, vhost string) (*exchangev1beta1.ExchangeObservation, error) {
	path := fmt.Sprintf("/api/exchanges/%s/%s", url.PathEscape(vhost), url.PathEscape(name))
	_, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return &exchangev1beta1.ExchangeObservation{Name: name, VHost: vhost}, nil
}

func (c *rabbitmqClient) CreateExchange(ctx context.Context, spec *exchangev1beta1.ExchangeParameters) (*exchangev1beta1.ExchangeObservation, error) {
	path := fmt.Sprintf("/api/exchanges/%s/%s", url.PathEscape(spec.VHost), url.PathEscape(spec.Name))
	body := map[string]interface{}{
		"type": spec.Type,
	}
	_, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}
	return &exchangev1beta1.ExchangeObservation{Name: spec.Name, VHost: spec.VHost}, nil
}

func (c *rabbitmqClient) DeleteExchange(ctx context.Context, name, vhost string) error {
	path := fmt.Sprintf("/api/exchanges/%s/%s", url.PathEscape(vhost), url.PathEscape(name))
	_, err := c.request(ctx, "DELETE", path, nil)
	return err
}

func (c *rabbitmqClient) GetQueue(ctx context.Context, name, vhost string) (*queuev1beta1.QueueObservation, error) {
	path := fmt.Sprintf("/api/queues/%s/%s", url.PathEscape(vhost), url.PathEscape(name))
	_, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return &queuev1beta1.QueueObservation{Name: name, VHost: vhost}, nil
}

func (c *rabbitmqClient) CreateQueue(ctx context.Context, spec *queuev1beta1.QueueParameters) (*queuev1beta1.QueueObservation, error) {
	path := fmt.Sprintf("/api/queues/%s/%s", url.PathEscape(spec.VHost), url.PathEscape(spec.Name))
	body := map[string]interface{}{
		"durable": spec.Durable,
	}
	_, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}
	return &queuev1beta1.QueueObservation{Name: spec.Name, VHost: spec.VHost}, nil
}

func (c *rabbitmqClient) DeleteQueue(ctx context.Context, name, vhost string) error {
	path := fmt.Sprintf("/api/queues/%s/%s", url.PathEscape(vhost), url.PathEscape(name))
	_, err := c.request(ctx, "DELETE", path, nil)
	return err
}

func (c *rabbitmqClient) GetBinding(ctx context.Context, source, destination, vhost, routingKey string) (*bindingv1beta1.BindingObservation, error) {
	path := fmt.Sprintf("/api/bindings/%s/e/%s/q/%s", url.PathEscape(vhost), url.PathEscape(source), url.PathEscape(destination))
	data, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var entries []struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		VHost       string `json:"vhost"`
		RoutingKey  string `json:"routing_key"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal bindings")
	}
	for _, b := range entries {
		if b.RoutingKey == routingKey {
			return &bindingv1beta1.BindingObservation{
				Source:      b.Source,
				Destination: b.Destination,
				VHost:       b.VHost,
				RoutingKey:  b.RoutingKey,
			}, nil
		}
	}
	return nil, &NotFoundError{Body: "binding not found"}
}

func (c *rabbitmqClient) CreateBinding(ctx context.Context, spec *bindingv1beta1.BindingParameters) (*bindingv1beta1.BindingObservation, error) {
	path := fmt.Sprintf("/api/bindings/%s/e/%s/q/%s", url.PathEscape(spec.VHost), url.PathEscape(spec.Source), url.PathEscape(spec.Destination))
	body := map[string]interface{}{
		"routing_key": spec.RoutingKey,
	}
	_, err := c.request(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}
	return &bindingv1beta1.BindingObservation{
		Source:      spec.Source,
		Destination: spec.Destination,
		VHost:       spec.VHost,
		RoutingKey:  spec.RoutingKey,
	}, nil
}

func (c *rabbitmqClient) DeleteBinding(ctx context.Context, source, destination, vhost, routingKey string) error {
	path := fmt.Sprintf("/api/bindings/%s/e/%s/q/%s/%s", url.PathEscape(vhost), url.PathEscape(source), url.PathEscape(destination), url.PathEscape(routingKey))
	_, err := c.request(ctx, "DELETE", path, nil)
	return err
}

func (c *rabbitmqClient) GetUser(ctx context.Context, name string) (*userv1beta1.UserObservation, error) {
	path := fmt.Sprintf("/api/users/%s", url.PathEscape(name))
	data, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var u struct {
		Name string `json:"name"`
		Tags string `json:"tags"`
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal user")
	}
	obs := &userv1beta1.UserObservation{Name: u.Name}
	if u.Tags != "" {
		obs.Tags = strings.Split(u.Tags, ",")
	}
	return obs, nil
}

func (c *rabbitmqClient) CreateUser(ctx context.Context, spec *userv1beta1.UserParameters, password string) (*userv1beta1.UserObservation, error) {
	path := fmt.Sprintf("/api/users/%s", url.PathEscape(spec.Name))
	body := map[string]interface{}{
		"password": password,
		"tags":     strings.Join(spec.Tags, ","),
	}
	_, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}
	return &userv1beta1.UserObservation{Name: spec.Name}, nil
}

func (c *rabbitmqClient) DeleteUser(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/users/%s", url.PathEscape(name))
	_, err := c.request(ctx, "DELETE", path, nil)
	return err
}

func (c *rabbitmqClient) GetPermission(ctx context.Context, user, vhost string) (*permissionv1beta1.PermissionObservation, error) {
	path := fmt.Sprintf("/api/permissions/%s/%s", url.PathEscape(vhost), url.PathEscape(user))
	data, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var p struct {
		User      string `json:"user"`
		VHost     string `json:"vhost"`
		Configure string `json:"configure"`
		Write     string `json:"write"`
		Read      string `json:"read"`
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal permission")
	}
	return &permissionv1beta1.PermissionObservation{
		User:      p.User,
		VHost:     p.VHost,
		Configure: p.Configure,
		Write:     p.Write,
		Read:      p.Read,
	}, nil
}

func (c *rabbitmqClient) SetPermission(ctx context.Context, spec *permissionv1beta1.PermissionParameters) (*permissionv1beta1.PermissionObservation, error) {
	path := fmt.Sprintf("/api/permissions/%s/%s", url.PathEscape(spec.VHost), url.PathEscape(spec.User))
	body := map[string]interface{}{
		"configure": spec.Configure,
		"write":     spec.Write,
		"read":      spec.Read,
	}
	_, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}
	return &permissionv1beta1.PermissionObservation{User: spec.User, VHost: spec.VHost}, nil
}

func (c *rabbitmqClient) DeletePermission(ctx context.Context, user, vhost string) error {
	path := fmt.Sprintf("/api/permissions/%s/%s", url.PathEscape(vhost), url.PathEscape(user))
	_, err := c.request(ctx, "DELETE", path, nil)
	return err
}
