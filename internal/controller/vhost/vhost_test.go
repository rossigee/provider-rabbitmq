package vhost

import (
	permissionv1beta1 "github.com/rossigee/provider-rabbitmq/apis/permission/v1beta1"
	"context"
	"testing"

	"github.com/pkg/errors"
	bindingv1beta1 "github.com/rossigee/provider-rabbitmq/apis/binding/v1beta1"
	exchangev1beta1 "github.com/rossigee/provider-rabbitmq/apis/exchange/v1beta1"
	queuev1beta1 "github.com/rossigee/provider-rabbitmq/apis/queue/v1beta1"
	userv1beta1 "github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
	vhostv1beta1 "github.com/rossigee/provider-rabbitmq/apis/vhost/v1beta1"
	"github.com/rossigee/provider-rabbitmq/internal/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopClient satisfies clients.Client with zero-value returns.
type noopClient struct{}

func (*noopClient) GetVhost(_ context.Context, _ string) (*vhostv1beta1.VhostObservation, error) {
	return nil, nil
}
func (*noopClient) CreateVhost(_ context.Context, _ *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error) {
	return &vhostv1beta1.VhostObservation{}, nil
}
func (*noopClient) DeleteVhost(_ context.Context, _ string) error { return nil }
func (*noopClient) GetExchange(_ context.Context, _, _ string) (*exchangev1beta1.ExchangeObservation, error) {
	return nil, nil
}
func (*noopClient) CreateExchange(_ context.Context, _ *exchangev1beta1.ExchangeParameters) (*exchangev1beta1.ExchangeObservation, error) {
	return nil, nil
}
func (*noopClient) DeleteExchange(_ context.Context, _, _ string) error { return nil }
func (*noopClient) GetQueue(_ context.Context, _, _ string) (*queuev1beta1.QueueObservation, error) {
	return nil, nil
}
func (*noopClient) CreateQueue(_ context.Context, _ *queuev1beta1.QueueParameters) (*queuev1beta1.QueueObservation, error) {
	return nil, nil
}
func (*noopClient) DeleteQueue(_ context.Context, _, _ string) error { return nil }
func (*noopClient) GetBinding(_ context.Context, _, _, _, _ string) (*bindingv1beta1.BindingObservation, error) {
	return nil, nil
}
func (*noopClient) CreateBinding(_ context.Context, _ *bindingv1beta1.BindingParameters) (*bindingv1beta1.BindingObservation, error) {
	return nil, nil
}
func (*noopClient) DeleteBinding(_ context.Context, _, _, _, _ string) error { return nil }
func (*noopClient) GetUser(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
	return nil, nil
}
func (*noopClient) CreateUser(_ context.Context, _ *userv1beta1.UserParameters, _ string) (*userv1beta1.UserObservation, error) {
	return nil, nil
}
func (*noopClient) DeleteUser(_ context.Context, _ string) error { return nil }
func (*noopClient) GetPermission(_ context.Context, _, _ string) (*permissionv1beta1.PermissionObservation, error) {
	return nil, nil
}
func (*noopClient) SetPermission(_ context.Context, _ *permissionv1beta1.PermissionParameters) (*permissionv1beta1.PermissionObservation, error) {
	return nil, nil
}
func (*noopClient) DeletePermission(_ context.Context, _, _ string) error { return nil }

// vhostStub overrides just the three vhost methods.
type vhostStub struct {
	noopClient
	getVhost    func(context.Context, string) (*vhostv1beta1.VhostObservation, error)
	createVhost func(context.Context, *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error)
	deleteVhost func(context.Context, string) error
}

func (s *vhostStub) GetVhost(ctx context.Context, name string) (*vhostv1beta1.VhostObservation, error) {
	if s.getVhost != nil {
		return s.getVhost(ctx, name)
	}
	return s.noopClient.GetVhost(ctx, name)
}
func (s *vhostStub) CreateVhost(ctx context.Context, spec *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error) {
	if s.createVhost != nil {
		return s.createVhost(ctx, spec)
	}
	return s.noopClient.CreateVhost(ctx, spec)
}
func (s *vhostStub) DeleteVhost(ctx context.Context, name string) error {
	if s.deleteVhost != nil {
		return s.deleteVhost(ctx, name)
	}
	return s.noopClient.DeleteVhost(ctx, name)
}

var _ clients.Client = &noopClient{}
var _ clients.Client = &vhostStub{}

// --- helpers ---

func newVhost(name string) *vhostv1beta1.Vhost {
	return &vhostv1beta1.Vhost{
		Spec: vhostv1beta1.VhostSpec{
			ForProvider: vhostv1beta1.VhostParameters{Name: name},
		},
	}
}

// --- Observe ---

func TestObserve_NotFound(t *testing.T) {
	e := &external{service: &vhostStub{
		getVhost: func(_ context.Context, _ string) (*vhostv1beta1.VhostObservation, error) {
			return nil, &clients.NotFoundError{}
		},
	}}
	obs, err := e.Observe(context.Background(), newVhost("staging"))
	require.NoError(t, err)
	assert.False(t, obs.ResourceExists)
}

func TestObserve_Exists(t *testing.T) {
	e := &external{service: &vhostStub{
		getVhost: func(_ context.Context, name string) (*vhostv1beta1.VhostObservation, error) {
			return &vhostv1beta1.VhostObservation{Name: name}, nil
		},
	}}
	obs, err := e.Observe(context.Background(), newVhost("staging"))
	require.NoError(t, err)
	assert.True(t, obs.ResourceExists)
	// Vhosts are immutable in RabbitMQ; always reported as up-to-date.
	assert.True(t, obs.ResourceUpToDate)
}

func TestObserve_ErrorPropagated(t *testing.T) {
	e := &external{service: &vhostStub{
		getVhost: func(_ context.Context, _ string) (*vhostv1beta1.VhostObservation, error) {
			return nil, errors.New("connection refused")
		},
	}}
	_, err := e.Observe(context.Background(), newVhost("staging"))
	require.Error(t, err)
}

// --- Create ---

func TestCreate_CallsCreateVhost(t *testing.T) {
	var gotSpec *vhostv1beta1.VhostParameters
	e := &external{service: &vhostStub{
		createVhost: func(_ context.Context, spec *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error) {
			gotSpec = spec
			return &vhostv1beta1.VhostObservation{Name: spec.Name}, nil
		},
	}}
	_, err := e.Create(context.Background(), newVhost("staging"))
	require.NoError(t, err)
	require.NotNil(t, gotSpec)
	assert.Equal(t, "staging", gotSpec.Name)
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	var deletedName string
	e := &external{service: &vhostStub{
		deleteVhost: func(_ context.Context, name string) error {
			deletedName = name
			return nil
		},
	}}
	_, err := e.Delete(context.Background(), newVhost("staging"))
	require.NoError(t, err)
	assert.Equal(t, "staging", deletedName)
}

func TestDelete_AlreadyGoneIsIgnored(t *testing.T) {
	e := &external{service: &vhostStub{
		deleteVhost: func(_ context.Context, _ string) error {
			return &clients.NotFoundError{}
		},
	}}
	_, err := e.Delete(context.Background(), newVhost("staging"))
	require.NoError(t, err)
}

func TestDelete_ErrorPropagated(t *testing.T) {
	e := &external{service: &vhostStub{
		deleteVhost: func(_ context.Context, _ string) error {
			return errors.New("timeout")
		},
	}}
	_, err := e.Delete(context.Background(), newVhost("staging"))
	require.Error(t, err)
}
