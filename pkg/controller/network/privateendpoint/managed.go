/*
Copyright 2019 The Crossplane Authors.

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

package privateendpoint

import (
	"context"

	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network/networkapi"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/network/v1beta1"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network"
)

// Error strings.
const (
	errNotPrivateEndpoint    = "managed resource is not a Private Endpoint"
	errCreatePrivateEndpoint = "cannot create Private Endpoint"
	errUpdatePrivateEndpoint = "cannot update Private Endpoint"
	errGetPrivateEndpoint    = "cannot get Private Endpoint"
	errDeletePrivateEndpoint = "cannot delete Private Endpoint"
)

// Setup adds a controller that reconciles Subnets.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.PrivateEndpointKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.PrivateEndpoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.PrivateEndpointGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := azurenetwork.NewPrivateEndpointsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client networkapi.PrivateEndpointsClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	pe, ok := mg.(*v1beta1.PrivateEndpoint)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPrivateEndpoint)
	}

	az, err := e.client.Get(ctx, pe.Spec.ResourceGroupName, meta.GetExternalName(pe), "")
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPrivateEndpoint)
	}

	network.UpdatePrivateEndpointStatusFromAzure(pe, az)
	pe.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	pe, ok := mg.(*v1beta1.PrivateEndpoint)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPrivateEndpoint)
	}

	pe.Status.SetConditions(xpv1.Creating())

	endpoint := network.NewPrivateEndpointParameters(pe)
	if _, err := e.client.CreateOrUpdate(ctx, pe.Spec.ResourceGroupName, meta.GetExternalName(pe), endpoint); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePrivateEndpoint)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	pe, ok := mg.(*v1beta1.PrivateEndpoint)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPrivateEndpoint)
	}

	az, err := e.client.Get(ctx, pe.Spec.ResourceGroupName, meta.GetExternalName(pe), "")
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetPrivateEndpoint)
	}

	if network.PrivateEndpointNeedsUpdate(pe, az) {
		snet := network.NewPrivateEndpointParameters(pe)
		if _, err := e.client.CreateOrUpdate(ctx, pe.Spec.ResourceGroupName, meta.GetExternalName(pe), snet); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePrivateEndpoint)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	pe, ok := mg.(*v1beta1.PrivateEndpoint)
	if !ok {
		return errors.New(errNotPrivateEndpoint)
	}

	mg.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, pe.Spec.ResourceGroupName, meta.GetExternalName(pe))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeletePrivateEndpoint)
}
