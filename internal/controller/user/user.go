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

package user

import (
	"context"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-rabbitmq/apis/user/v1beta1"
	clients "github.com/rossigee/provider-rabbitmq/internal/clients"
)

const errResolvePassword = "cannot resolve password secret"

const errNotUser = "managed resource is not a User custom resource"

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.UserKind)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.UserGroupVersionKind),
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newServiceFn: clients.NewClient}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(nil))
	return ctrl.NewControllerManagedBy(mgr).Named(name).WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).For(&v1beta1.User{}).Complete(r)
}

type connector struct {
	kube         client.Client
	newServiceFn func(config *clients.Config) clients.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get config")
	}
	return &external{service: c.newServiceFn(config), kube: c.kube}, nil
}

type external struct {
	service clients.Client
	kube    client.Client
}

func (c *external) Disconnect(ctx context.Context) error { return nil }

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.User)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotUser)
	}
	user, err := c.service.GetUser(ctx, cr.Spec.ForProvider.Name)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get user")
	}
	cr.Status.AtProvider = *user
	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: equalStringSlices(cr.Spec.ForProvider.Tags, user.Tags),
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.User)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotUser)
	}

	password := ""
	if cr.Spec.ForProvider.PasswordSecretRef != nil {
		data, err := resource.CommonCredentialExtractor(
			ctx,
			xpv1.CredentialsSourceSecret,
			c.kube,
			xpv1.CommonCredentialSelectors{SecretRef: cr.Spec.ForProvider.PasswordSecretRef},
		)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errResolvePassword)
		}
		password = string(data)
	}

	cr.SetConditions(xpv1.Creating())
	user, err := c.service.CreateUser(ctx, &cr.Spec.ForProvider, password)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create user")
	}
	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
	cr.Status.AtProvider = *user
	cr.SetConditions(xpv1.Available())
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.User)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotUser)
	}
	password := ""
	if cr.Spec.ForProvider.PasswordSecretRef != nil {
		data, err := resource.CommonCredentialExtractor(
			ctx,
			xpv1.CredentialsSourceSecret,
			c.kube,
			xpv1.CommonCredentialSelectors{SecretRef: cr.Spec.ForProvider.PasswordSecretRef},
		)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errResolvePassword)
		}
		password = string(data)
	}
	_, err := c.service.CreateUser(ctx, &cr.Spec.ForProvider, password)
	return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update user")
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.User)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotUser)
	}
	cr.SetConditions(xpv1.Deleting())
	err := c.service.DeleteUser(ctx, cr.Spec.ForProvider.Name)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete user")
	}
	return managed.ExternalDelete{}, nil
}
