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
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	"github.com/crossplane/provider-azure/apis/network/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network/fake"
)

const (
	name              = "coolSubnet"
	uid               = types.UID("definitely-a-uuid")
	addressPrefix     = "10.0.0.0/16"
	vnetID            = "coolVnet"
	privateLinkName   = "coolPrivateLink"
	privateLinkID     = "coolPrivateLink-uuid"
	resourceGroupName = "coolRG"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")
)

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

type privateEndpointModifier func(*v1beta1.PrivateEndpoint)

func withConditions(c ...xpv1.Condition) privateEndpointModifier {
	return func(r *v1beta1.PrivateEndpoint) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) privateEndpointModifier {
	return func(r *v1beta1.PrivateEndpoint) { r.Status.State = s }
}

func privateendpoint(sm ...privateEndpointModifier) *v1beta1.PrivateEndpoint {
	r := &v1beta1.PrivateEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1beta1.PrivateEndpointSpec{
			ResourceGroupName: resourceGroupName,
			PrivateEndpointParameters: v1beta1.PrivateEndpointParameters{
				VirtualNetworkSubnetID: vnetID,
				PrivateLinkServiceConnections: []v1beta1.PrivateLinkServiceConnection{
					{
						Name:                        privateLinkName,
						PrivateConnectionResourceID: privateLinkID,
						SubresourceID:               []string{},
					},
				},
				ManualPrivateLinkServiceConnections: []v1beta1.PrivateLinkServiceConnection{},
			},
		},
		Status: v1beta1.PrivateEndpointStatus{},
	}

	meta.SetExternalName(r, name)

	for _, m := range sm {
		m(r)
	}

	return r
}

// Test that our Reconciler implementation satisfies the Reconciler interface.
var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (network.PrivateEndpointsCreateOrUpdateFuture, error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, nil
				},
			}},
			r: privateendpoint(),
			want: privateendpoint(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (network.PrivateEndpointsCreateOrUpdateFuture, error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: privateendpoint(),
			want: privateendpoint(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreatePrivateEndpoint),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Create(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	cases := []testCase{
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
							PrivateEndpointProperties: &network.PrivateEndpointProperties{
								Subnet: &network.Subnet{},
							},
						}, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
				},
			}},
			r:    privateendpoint(),
			want: privateendpoint(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							NetworkInterfaces: &[]network.Interface{},
							Subnet:            &network.Subnet{},
							ProvisioningState: network.ProvisioningState("Available"),
						},
					}, nil
				},
			}},
			r: privateendpoint(),
			want: privateendpoint(
				withConditions(xpv1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{}, errorBoom
				},
			}},
			r:       privateendpoint(),
			want:    privateendpoint(),
			wantErr: errors.Wrap(errorBoom, errGetPrivateEndpoint),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Observe(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPrivateEndpoint",
			e:       &external{client: &fake.MockPrivateEndpointsClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotPrivateEndpoint),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{
								ID: azure.ToStringPtr(vnetID),
							},
							PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
								{
									Name: azure.ToStringPtr(privateLinkName),
									ID:   azure.ToStringPtr(privateLinkID),
								},
							},
							ManualPrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{},
						},
					}, nil
				},
			}},
			r:    privateendpoint(),
			want: privateendpoint(),
		},
		{
			name: "SuccessfulNeedsUpdate",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{},
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (network.PrivateEndpointsCreateOrUpdateFuture, error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, nil
				},
			}},
			r:    privateendpoint(),
			want: privateendpoint(),
		},
		{
			name: "UnsuccessfulGet",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{},
						},
					}, errorBoom
				},
			}},
			r:       privateendpoint(),
			want:    privateendpoint(),
			wantErr: errors.Wrap(errorBoom, errGetPrivateEndpoint),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockPrivateEndpointsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{},
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (network.PrivateEndpointsCreateOrUpdateFuture, error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       privateendpoint(),
			want:    privateendpoint(),
			wantErr: errors.Wrap(errorBoom, errUpdatePrivateEndpoint),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Update(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPrivateEndpoint",
			e:       &managed.NopClient{},
			r:       &v1beta1.PrivateEndpoint{},
			want:    &v1beta1.PrivateEndpoint{},
			wantErr: errors.New(errNotPrivateEndpoint),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.e.Delete(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
