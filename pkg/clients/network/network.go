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

package network

import (
	"reflect"

	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	"github.com/crossplane/provider-azure/apis/network/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// NewVirtualNetworkParameters returns an Azure VirtualNetwork object from a virtual network spec
func NewVirtualNetworkParameters(v *v1alpha3.VirtualNetwork) networkmgmt.VirtualNetwork {
	return networkmgmt.VirtualNetwork{
		Location: azure.ToStringPtr(v.Spec.Location),
		Tags:     azure.ToStringPtrMap(v.Spec.Tags),
		VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
			EnableDdosProtection: azure.ToBoolPtr(v.Spec.VirtualNetworkPropertiesFormat.EnableDDOSProtection, azure.FieldRequired),
			EnableVMProtection:   azure.ToBoolPtr(v.Spec.VirtualNetworkPropertiesFormat.EnableVMProtection, azure.FieldRequired),
			AddressSpace: &networkmgmt.AddressSpace{
				AddressPrefixes: &v.Spec.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes,
			},
		},
	}
}

// VirtualNetworkNeedsUpdate determines if a virtual network need to be updated
func VirtualNetworkNeedsUpdate(kube *v1alpha3.VirtualNetwork, az networkmgmt.VirtualNetwork) bool {
	up := NewVirtualNetworkParameters(kube)

	switch {
	case !reflect.DeepEqual(up.VirtualNetworkPropertiesFormat.AddressSpace, az.VirtualNetworkPropertiesFormat.AddressSpace):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkPropertiesFormat.EnableDdosProtection, az.VirtualNetworkPropertiesFormat.EnableDdosProtection):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkPropertiesFormat.EnableVMProtection, az.VirtualNetworkPropertiesFormat.EnableVMProtection):
		return true
	case !reflect.DeepEqual(up.Tags, az.Tags):
		return true
	}

	return false
}

// UpdateVirtualNetworkStatusFromAzure updates the status related to the external
// Azure virtual network in the VirtualNetworkStatus
func UpdateVirtualNetworkStatusFromAzure(v *v1alpha3.VirtualNetwork, az networkmgmt.VirtualNetwork) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ResourceGUID = azure.ToString(az.ResourceGUID)
	v.Status.Type = azure.ToString(az.Type)
}

// NewSubnetParameters returns an Azure Subnet object from a subnet spec
func NewSubnetParameters(s *v1alpha3.Subnet) networkmgmt.Subnet {
	return networkmgmt.Subnet{
		SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
			AddressPrefix:    azure.ToStringPtr(s.Spec.SubnetPropertiesFormat.AddressPrefix),
			ServiceEndpoints: NewServiceEndpoints(s.Spec.SubnetPropertiesFormat.ServiceEndpoints),
		},
	}
}

// NewServiceEndpoints converts to Azure ServiceEndpointPropertiesFormat
func NewServiceEndpoints(e []v1alpha3.ServiceEndpointPropertiesFormat) *[]networkmgmt.ServiceEndpointPropertiesFormat {
	endpoints := make([]networkmgmt.ServiceEndpointPropertiesFormat, len(e))

	for i, end := range e {
		endpoints[i] = networkmgmt.ServiceEndpointPropertiesFormat{
			Service: azure.ToStringPtr(end.Service),
		}
	}

	return &endpoints
}

// SubnetNeedsUpdate determines if a virtual network need to be updated
func SubnetNeedsUpdate(kube *v1alpha3.Subnet, az networkmgmt.Subnet) bool {
	up := NewSubnetParameters(kube)

	return !reflect.DeepEqual(up.SubnetPropertiesFormat.AddressPrefix, az.SubnetPropertiesFormat.AddressPrefix)
}

// UpdateSubnetStatusFromAzure updates the status related to the external
// Azure subnet in the SubnetStatus
func UpdateSubnetStatusFromAzure(v *v1alpha3.Subnet, az networkmgmt.Subnet) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Purpose = azure.ToString(az.Purpose)
}

// NewprivateLinkServiceConnections converts to Azure PrivateLinkServiceConnection
func NewprivateLinkServiceConnections(p []v1beta1.PrivateLinkServiceConnection) *[]networkmgmt.PrivateLinkServiceConnection {
	connections := make([]networkmgmt.PrivateLinkServiceConnection, len(p))

	for i, conn := range p {
		connections[i] = networkmgmt.PrivateLinkServiceConnection{
			Name: azure.ToStringPtr(conn.Name),
			PrivateLinkServiceConnectionProperties: &networkmgmt.PrivateLinkServiceConnectionProperties{
				PrivateLinkServiceID: azure.ToStringPtr(conn.PrivateConnectionResourceID),
				GroupIds:             azure.ToStringArrayPtr(conn.SubresourceIDs),
			},
		}
	}

	return &connections
}

// NewPrivateEndpointParameters returns an Azure Private Endpoint object from a private endpoint spec
func NewPrivateEndpointParameters(v *v1beta1.PrivateEndpoint) networkmgmt.PrivateEndpoint {
	return networkmgmt.PrivateEndpoint{
		Location: azure.ToStringPtr(v.Spec.Location),
		Tags:     azure.ToStringPtrMap(v.Spec.Tags),
		PrivateEndpointProperties: &networkmgmt.PrivateEndpointProperties{
			Subnet:                              &networkmgmt.Subnet{ID: azure.ToStringPtr(v.Spec.VirtualNetworkSubnetID)},
			PrivateLinkServiceConnections:       NewprivateLinkServiceConnections(v.Spec.PrivateLinkServiceConnections),
			ManualPrivateLinkServiceConnections: NewprivateLinkServiceConnections(v.Spec.ManualPrivateLinkServiceConnections),
		},
	}
}

// PrivateEndpointNeedsUpdate determines if a private endpoint need to be updated
func PrivateEndpointNeedsUpdate(kube *v1beta1.PrivateEndpoint, az networkmgmt.PrivateEndpoint) bool {
	up := NewPrivateEndpointParameters(kube)

	switch {
	case !cmp.Equal(up.PrivateEndpointProperties.Subnet, az.PrivateEndpointProperties.Subnet):
		return true
	case !reflect.DeepEqual(up.PrivateEndpointProperties.ManualPrivateLinkServiceConnections, az.PrivateEndpointProperties.ManualPrivateLinkServiceConnections):
		return true
	case !cmp.Equal(up.PrivateEndpointProperties.PrivateLinkServiceConnections, az.PrivateEndpointProperties.PrivateLinkServiceConnections):
		return true
	case !reflect.DeepEqual(up.Tags, az.Tags):
		return true
	}

	return false
}

// UpdatePrivateEndpointStatusFromAzure updates the status related to the external
// Azure private endpoint in the PrivateEndpointStatus
func UpdatePrivateEndpointStatusFromAzure(v *v1beta1.PrivateEndpoint, az networkmgmt.PrivateEndpoint) {
	networkInterfaces := make([]string, len(*az.NetworkInterfaces))
	for i, n := range *az.NetworkInterfaces {
		networkInterfaces[i] = azure.ToString(n.ID)
	}

	v.Status.State = string(az.ProvisioningState)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.Type = azure.ToString(az.Type)
	v.Status.NetworkInterfacesID = networkInterfaces
}
