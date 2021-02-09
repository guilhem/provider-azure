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

package v1beta1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	netv1alpha3 "github.com/crossplane/provider-azure/apis/network/v1alpha3"
	"github.com/crossplane/provider-azure/apis/v1alpha3"
)

// ResolveReferences of PrivateEndpoint
func (mg *PrivateEndpoint) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ResourceGroupName,
		Reference:    mg.Spec.ResourceGroupNameRef,
		Selector:     mg.Spec.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.resourceGroupName")
	}
	mg.Spec.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ResourceGroupNameRef = rsp.ResolvedReference

	// Resolve spec.virtualNetworkSubnetId
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.VirtualNetworkSubnetID,
		Reference:    mg.Spec.VirtualNetworkSubnetIDRef,
		Selector:     mg.Spec.VirtualNetworkSubnetIDSelector,
		To:           reference.To{Managed: &netv1alpha3.Subnet{}, List: &netv1alpha3.SubnetList{}},
		Extract:      netv1alpha3.SubnetID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.virtualNetworkSubnetId")
	}
	mg.Spec.VirtualNetworkSubnetID = rsp.ResolvedValue
	mg.Spec.VirtualNetworkSubnetIDRef = rsp.ResolvedReference

	return nil
}
