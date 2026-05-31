package user

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"

	bindingv1beta1 "github.com/rossigee/provider-rabbitmq/apis/binding/v1beta1"
	exchangev1beta1 "github.com/rossigee/provider-rabbitmq/apis/exchange/v1beta1"
	permv1beta1 "github.com/rossigee/provider-rabbitmq/apis/permission/v1beta1"
	queuev1beta1 "github.com/rossigee/provider-rabbitmq/apis/queue/v1beta1"
	userv1beta1 "github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
	vhostv1beta1 "github.com/rossigee/provider-rabbitmq/apis/vhost/v1beta1"
	clients "github.com/rossigee/provider-rabbitmq/internal/clients"
)

// noopClient satisfies clients.Client with zero-value returns.
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
func (*noopClient) GetPermission(_ context.Context, _, _ string) (*permv1beta1.PermissionObservation, error) {
	return nil, nil
}
func (*noopClient) SetPermission(_ context.Context, _ *permv1beta1.PermissionParameters) (*permv1beta1.PermissionObservation, error) {
	return nil, nil
}
func (*noopClient) DeletePermission(_ context.Context, _, _ string) error { return nil }
func (*noopClient) GetUser(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
	return nil, nil
}
func (*noopClient) CreateUser(_ context.Context, _ *userv1beta1.UserParameters, _ string) (*userv1beta1.UserObservation, error) {
	return &userv1beta1.UserObservation{}, nil
}
func (*noopClient) DeleteUser(_ context.Context, _ string) error { return nil }

// userStub overrides just the three user methods.
type userStub struct {
	noopClient
	getUser    func(context.Context, string) (*userv1beta1.UserObservation, error)
	createUser func(context.Context, *userv1beta1.UserParameters, string) (*userv1beta1.UserObservation, error)
	deleteUser func(context.Context, string) error
}

func (s *userStub) GetUser(ctx context.Context, name string) (*userv1beta1.UserObservation, error) {
	if s.getUser != nil {
		return s.getUser(ctx, name)
	}
	return s.noopClient.GetUser(ctx, name)
}
func (s *userStub) CreateUser(ctx context.Context, spec *userv1beta1.UserParameters, password string) (*userv1beta1.UserObservation, error) {
	if s.createUser != nil {
		return s.createUser(ctx, spec, password)
	}
	return s.noopClient.CreateUser(ctx, spec, password)
}
func (s *userStub) DeleteUser(ctx context.Context, name string) error {
	if s.deleteUser != nil {
		return s.deleteUser(ctx, name)
	}
	return s.noopClient.DeleteUser(ctx, name)
}

var _ clients.Client = &noopClient{}
var _ clients.Client = &userStub{}

// --- equalStringSlices ---

func TestEqualStringSlices_Equal(t *testing.T) {
	assert.True(t, equalStringSlices([]string{"a", "b"}, []string{"a", "b"}))
	assert.True(t, equalStringSlices(nil, nil))
	assert.True(t, equalStringSlices([]string{}, []string{}))
}

func TestEqualStringSlices_DifferentLength(t *testing.T) {
	assert.False(t, equalStringSlices([]string{"a"}, []string{"a", "b"}))
}

func TestEqualStringSlices_DifferentContent(t *testing.T) {
	assert.False(t, equalStringSlices([]string{"a", "b"}, []string{"a", "c"}))
}

func TestEqualStringSlices_OrderMatters(t *testing.T) {
	assert.False(t, equalStringSlices([]string{"b", "a"}, []string{"a", "b"}))
}

// --- helpers ---

func newUser(name string, tags ...string) *userv1beta1.User {
	return &userv1beta1.User{
		Spec: userv1beta1.UserSpec{
			ForProvider: userv1beta1.UserParameters{
				Name: name,
				Tags: tags,
			},
		},
	}
}

func newExternalNoKube(svc clients.Client) *external {
	return &external{service: svc, kube: nil}
}

// --- Observe ---

func TestObserve_NotFound(t *testing.T) {
	e := newExternalNoKube(&userStub{
		getUser: func(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
			return nil, &clients.NotFoundError{}
		},
	})
	obs, err := e.Observe(context.Background(), newUser("alice", "administrator"))
	require.NoError(t, err)
	assert.False(t, obs.ResourceExists)
}

func TestObserve_TagsMatch(t *testing.T) {
	e := newExternalNoKube(&userStub{
		getUser: func(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
			return &userv1beta1.UserObservation{Name: "alice", Tags: []string{"administrator"}}, nil
		},
	})
	obs, err := e.Observe(context.Background(), newUser("alice", "administrator"))
	require.NoError(t, err)
	assert.True(t, obs.ResourceExists)
	assert.True(t, obs.ResourceUpToDate)
}

func TestObserve_TagsDiffer(t *testing.T) {
	e := newExternalNoKube(&userStub{
		getUser: func(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
			return &userv1beta1.UserObservation{Name: "alice", Tags: []string{"monitoring"}}, nil
		},
	})
	obs, err := e.Observe(context.Background(), newUser("alice", "administrator"))
	require.NoError(t, err)
	assert.True(t, obs.ResourceExists)
	assert.False(t, obs.ResourceUpToDate)
}

func TestObserve_NoTags(t *testing.T) {
	e := newExternalNoKube(&userStub{
		getUser: func(_ context.Context, _ string) (*userv1beta1.UserObservation, error) {
			return &userv1beta1.UserObservation{Name: "alice"}, nil
		},
	})
	obs, err := e.Observe(context.Background(), newUser("alice"))
	require.NoError(t, err)
	assert.True(t, obs.ResourceUpToDate)
}

// --- Create (no password) ---

func TestCreate_NoPassword(t *testing.T) {
	var gotSpec *userv1beta1.UserParameters
	var gotPassword string
	e := newExternalNoKube(&userStub{
		createUser: func(_ context.Context, spec *userv1beta1.UserParameters, pw string) (*userv1beta1.UserObservation, error) {
			gotSpec, gotPassword = spec, pw
			return &userv1beta1.UserObservation{Name: spec.Name}, nil
		},
	})
	cr := newUser("alice", "administrator")
	_, err := e.Create(context.Background(), cr)
	require.NoError(t, err)
	require.NotNil(t, gotSpec)
	assert.Equal(t, "alice", gotSpec.Name)
	assert.Equal(t, "", gotPassword)
}

// --- Create (with password secret) ---

func TestCreate_WithPassword(t *testing.T) {
	s := kruntime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "alice-pass", Namespace: "default"},
		Data:       map[string][]byte{"password": []byte("s3cr3t")},
	}
	fakeKube := fake.NewClientBuilder().WithScheme(s).WithObjects(secret).Build()

	var gotPassword string
	e := &external{
		service: &userStub{
			createUser: func(_ context.Context, _ *userv1beta1.UserParameters, pw string) (*userv1beta1.UserObservation, error) {
				gotPassword = pw
				return &userv1beta1.UserObservation{}, nil
			},
		},
		kube: fakeKube,
	}

	cr := newUser("alice")
	cr.Spec.ForProvider.PasswordSecretRef = &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Name: "alice-pass", Namespace: "default"},
		Key:             "password",
	}
	_, err := e.Create(context.Background(), cr)
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", gotPassword)
}

// --- Update (no password) ---

func TestUpdate_NoPassword(t *testing.T) {
	var called bool
	e := newExternalNoKube(&userStub{
		createUser: func(_ context.Context, spec *userv1beta1.UserParameters, pw string) (*userv1beta1.UserObservation, error) {
			called = true
			assert.Equal(t, []string{"monitoring"}, spec.Tags)
			return &userv1beta1.UserObservation{}, nil
		},
	})
	_, err := e.Update(context.Background(), newUser("alice", "monitoring"))
	require.NoError(t, err)
	assert.True(t, called)
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	var deletedName string
	e := newExternalNoKube(&userStub{
		deleteUser: func(_ context.Context, name string) error {
			deletedName = name
			return nil
		},
	})
	_, err := e.Delete(context.Background(), newUser("alice"))
	require.NoError(t, err)
	assert.Equal(t, "alice", deletedName)
}

func TestDelete_AlreadyGoneIsIgnored(t *testing.T) {
	e := newExternalNoKube(&userStub{
		deleteUser: func(_ context.Context, _ string) error {
			return &clients.NotFoundError{}
		},
	})
	_, err := e.Delete(context.Background(), newUser("alice"))
	require.NoError(t, err)
}

func TestDelete_ErrorPropagated(t *testing.T) {
	e := newExternalNoKube(&userStub{
		deleteUser: func(_ context.Context, _ string) error {
			return errors.New("timeout")
		},
	})
	_, err := e.Delete(context.Background(), newUser("alice"))
	require.Error(t, err)
}
