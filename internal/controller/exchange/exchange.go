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

package exchange

import (
	"context"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-rabbitmq/apis/exchange/v1beta1"
	"github.com/rossigee/provider-rabbitmq/internal/clients"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const errNotExchange = "managed resource is not an Exchange custom resource"

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.ExchangeKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.ExchangeGroupVersionKind),
		managed.WithExternalConnector(&connector{kube: mgr.GetClient(), newServiceFn: clients.NewClient}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(nil))

	return ctrl.NewControllerManagedBy(mgr).Named(name).WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).For(&v1beta1.Exchange{}).Complete(r)
}

type connector struct {
	kube         client.Client
	newServiceFn func(config *clients.Config) clients.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	if _, ok := mg.(*v1beta1.Exchange); !ok {
		return nil, errors.New(errNotExchange)
	}
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get config")
	}
	return &external{service: c.newServiceFn(config)}, nil
}

type external struct{ service clients.Client }

func (c *external) Disconnect(ctx context.Context) error { return nil }

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Exchange)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotExchange)
	}
	exchange, err := c.service.GetExchange(ctx, cr.Spec.ForProvider.Name, cr.Spec.ForProvider.VHost)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get exchange")
	}
	cr.Status.AtProvider = *exchange
	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Exchange)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotExchange)
	}
	cr.SetConditions(xpv1.Creating())
	exchange, err := c.service.CreateExchange(ctx, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create exchange")
	}
	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
	cr.Status.AtProvider = *exchange
	cr.SetConditions(xpv1.Available())
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Exchange)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotExchange)
	}
	cr.SetConditions(xpv1.Deleting())
	err := c.service.DeleteExchange(ctx, cr.Spec.ForProvider.Name, cr.Spec.ForProvider.VHost)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete exchange")
	}
	return managed.ExternalDelete{}, nil
}
