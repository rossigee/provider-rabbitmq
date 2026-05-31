package permission

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"

	bindingv1beta1 "github.com/rossigee/provider-rabbitmq/apis/binding/v1beta1"
	exchangev1beta1 "github.com/rossigee/provider-rabbitmq/apis/exchange/v1beta1"
	permv1beta1 "github.com/rossigee/provider-rabbitmq/apis/permission/v1beta1"
	queuev1beta1 "github.com/rossigee/provider-rabbitmq/apis/queue/v1beta1"
	userv1beta1 "github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
	vhostv1beta1 "github.com/rossigee/provider-rabbitmq/apis/vhost/v1beta1"
	clients "github.com/rossigee/provider-rabbitmq/internal/clients"
)

// noopClient satisfies clients.Client with zero-value returns for all methods.
type noopClient struct{}

func (*noopClient) GetVhost(_ context.Context, _ string) (*vhostv1beta1.VhostObservation, error) {
	return nil, nil
}
func (*noopClient) CreateVhost(_ context.Context, _ *vhostv1beta1.VhostParameters) (*vhostv1beta1.VhostObservation, error) {
	return nil, nil
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
func (*noopClient) GetPermission(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
	return nil, nil
}
func (*noopClient) SetPermission(_ context.Context, _ *permv1beta1.PermissionParameters) (*permv1beta1.PermissionObservation, error) {
	return &permv1beta1.PermissionObservation{}, nil
}
func (*noopClient) DeletePermission(_ context.Context, _, _ string) error { return nil }

// permStub overrides just the three permission methods.
type permStub struct {
	noopClient
	getPermission    func(context.Context, string, string) (*permv1beta1.PermissionObservation, error)
	setPermission    func(context.Context, *permv1beta1.PermissionParameters) (*permv1beta1.PermissionObservation, error)
	deletePermission func(context.Context, string, string) error
}

func (s *permStub) GetPermission(ctx context.Context, user, vhost string) (*permv1beta1.PermissionObservation, error) {
	if s.getPermission != nil {
		return s.getPermission(ctx, user, vhost)
	}
	return s.noopClient.GetPermission(ctx, user, vhost)
}
func (s *permStub) SetPermission(ctx context.Context, spec *permv1beta1.PermissionParameters) (*permv1beta1.PermissionObservation, error) {
	if s.setPermission != nil {
		return s.setPermission(ctx, spec)
	}
	return s.noopClient.SetPermission(ctx, spec)
}
func (s *permStub) DeletePermission(ctx context.Context, user, vhost string) error {
	if s.deletePermission != nil {
		return s.deletePermission(ctx, user, vhost)
	}
	return s.noopClient.DeletePermission(ctx, user, vhost)
}

// Compile-time check that our stubs satisfy the interface.
var _ clients.Client = &noopClient{}
var _ clients.Client = &permStub{}

// --- helpers ---

func newPermission(user, vhost, configure, write, read string) *permv1beta1.Permission {
	return &permv1beta1.Permission{
		Spec: permv1beta1.PermissionSpec{
			ForProvider: permv1beta1.PermissionParameters{
				User: user, VHost: vhost,
				Configure: configure, Write: write, Read: read,
			},
		},
	}
}

// --- Observe ---

func TestObserve_NotFound(t *testing.T) {
	e := &external{service: &permStub{
		getPermission: func(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
			return nil, &clients.NotFoundError{Body: "not found"}
		},
	}}
	obs, err := e.Observe(context.Background(), newPermission("alice", "/", ".*", ".*", ".*"))
	require.NoError(t, err)
	assert.False(t, obs.ResourceExists)
}

func TestObserve_UpToDate(t *testing.T) {
	e := &external{service: &permStub{
		getPermission: func(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
			return &permv1beta1.PermissionObservation{Configure: ".*", Write: ".*", Read: ".*"}, nil
		},
	}}
	obs, err := e.Observe(context.Background(), newPermission("alice", "/", ".*", ".*", ".*"))
	require.NoError(t, err)
	assert.True(t, obs.ResourceExists)
	assert.True(t, obs.ResourceUpToDate)
}

func TestObserve_ConfigureDrift(t *testing.T) {
	e := &external{service: &permStub{
		getPermission: func(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
			return &permv1beta1.PermissionObservation{Configure: "old", Write: ".*", Read: ".*"}, nil
		},
	}}
	obs, err := e.Observe(context.Background(), newPermission("alice", "/", "new", ".*", ".*"))
	require.NoError(t, err)
	assert.True(t, obs.ResourceExists)
	assert.False(t, obs.ResourceUpToDate)
}

func TestObserve_ReadDrift(t *testing.T) {
	e := &external{service: &permStub{
		getPermission: func(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
			return &permv1beta1.PermissionObservation{Configure: ".*", Write: ".*", Read: "logs.*"}, nil
		},
	}}
	obs, err := e.Observe(context.Background(), newPermission("alice", "/", ".*", ".*", ".*"))
	require.NoError(t, err)
	assert.False(t, obs.ResourceUpToDate)
}

func TestObserve_ErrorPropagated(t *testing.T) {
	e := &external{service: &permStub{
		getPermission: func(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
			return nil, errors.New("connection refused")
		},
	}}
	_, err := e.Observe(context.Background(), newPermission("alice", "/", ".*", ".*", ".*"))
	require.Error(t, err)
}

// --- Create ---

func TestCreate_CallsSetPermission(t *testing.T) {
	var gotSpec *permv1beta1.PermissionParameters
	e := &external{service: &permStub{
		setPermission: func(_ context.Context, spec *permv1beta1.PermissionParameters) (*permv1beta1.PermissionObservation, error) {
			gotSpec = spec
			return &permv1beta1.PermissionObservation{User: spec.User, VHost: spec.VHost}, nil
		},
	}}
	cr := newPermission("alice", "/", ".*", ".*", ".*")
	_, err := e.Create(context.Background(), cr)
	require.NoError(t, err)
	require.NotNil(t, gotSpec)
	assert.Equal(t, "alice", gotSpec.User)
}

// --- Update ---

func TestUpdate_CallsSetPermission(t *testing.T) {
	var gotSpec *permv1beta1.PermissionParameters
	e := &external{service: &permStub{
		setPermission: func(_ context.Context, spec *permv1beta1.PermissionParameters) (*permv1beta1.PermissionObservation, error) {
			gotSpec = spec
			return &permv1beta1.PermissionObservation{}, nil
		},
	}}
	cr := newPermission("alice", "/", "newrule", ".*", ".*")
	_, err := e.Update(context.Background(), cr)
	require.NoError(t, err)
	require.NotNil(t, gotSpec)
	assert.Equal(t, "newrule", gotSpec.Configure)
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	var deletedUser, deletedVhost string
	e := &external{service: &permStub{
		deletePermission: func(_ context.Context, user, vhost string) error {
			deletedUser, deletedVhost = user, vhost
			return nil
		},
	}}
	_, err := e.Delete(context.Background(), newPermission("alice", "/", "", "", ""))
	require.NoError(t, err)
	assert.Equal(t, "alice", deletedUser)
	assert.Equal(t, "/", deletedVhost)
}

func TestDelete_AlreadyGoneIsIgnored(t *testing.T) {
	e := &external{service: &permStub{
		deletePermission: func(_ context.Context, _, _ string) error {
			return &clients.NotFoundError{}
		},
	}}
	_, err := e.Delete(context.Background(), newPermission("alice", "/", "", "", ""))
	require.NoError(t, err)
}

func TestDelete_ErrorPropagated(t *testing.T) {
	e := &external{service: &permStub{
		deletePermission: func(_ context.Context, _, _ string) error {
			return errors.New("network error")
		},
	}}
	_, err := e.Delete(context.Background(), newPermission("alice", "/", "", "", ""))
	require.Error(t, err)
}

// Compile-time check that external satisfies managed.ExternalClient.
var _ managed.ExternalClient = &external{}
