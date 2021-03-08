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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// PrivateLinkServiceConnection - defines properties of a Private Link Service Connection
type PrivateLinkServiceConnection struct {
	// Name - The AddressSpace that contains an array of IP address
	// ranges that can be used by subnets.
	Name string `json:"name"`

	// PrivateConnectionResourceID - The AddressSpace that contains an array of IP address
	// ranges that can be used by subnets.
	PrivateConnectionResourceID string `json:"privateConnectionResourceID"`

	// PrivateConnectionResourceIDRef - A reference to the the PrivateEndpoint's resource
	// group.
	PrivateConnectionResourceIDRef *xpv1.Reference `json:"privateConnectionResourceIDRef,omitempty"`

	// PrivateConnectionResourceIDSelector - Select a reference to the PrivateEndpoint's resource group.
	PrivateConnectionResourceIDSelector *xpv1.Selector `json:"privateConnectionResourceIDSelector,omitempty"`

	// SubresourceIDs - The AddressSpace that contains an array of IP address
	// ranges that can be used by subnets.
	SubresourceIDs []string `json:"subresourceIDs"`
}

// PrivateEndpointParameters - defines properties of a PrivateEndpoint.
type PrivateEndpointParameters struct {
	// VirtualNetworkSubnetID - The ARM resource id of the virtual network
	// subnet.
	VirtualNetworkSubnetID string `json:"virtualNetworkSubnetId,omitempty"`

	// VirtualNetworkSubnetIDRef - A reference to a Subnet to retrieve its ID
	VirtualNetworkSubnetIDRef *xpv1.Reference `json:"virtualNetworkSubnetIdRef,omitempty"`

	// VirtualNetworkSubnetIDRef - A selector for a Subnet to retrieve its ID
	VirtualNetworkSubnetIDSelector *xpv1.Selector `json:"virtualNetworkSubnetIdSelector,omitempty"`

	// PrivateLinkServiceConnections
	PrivateLinkServiceConnections []PrivateLinkServiceConnection `json:"privateLinkServiceConnections,omitempty"`

	// ManualPrivateLinkServiceConnections
	ManualPrivateLinkServiceConnections []PrivateLinkServiceConnection `json:"manualPrivateLinkServiceConnections,omitempty"`
}

// A PrivateEndpointSpec - defines the desired state of a PrivateEndpoint.
type PrivateEndpointSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ResourceGroupName - Name of the PrivateEndpoint's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the PrivateEndpoint's resource
	// group.
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to the PrivateEndpoint's resource group.
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// PrivateEndpointProperties - Properties of the PrivateEndpoint.
	PrivateEndpointParameters `json:",inline"`

	// Location - Resource location.
	Location string `json:"location"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// A PrivateEndpointStatus represents the observed state of the PrivateEndpoint
type PrivateEndpointStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// State of this PrivateEndpoint.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this PrivateEndpoint, if
	// any.
	Message string `json:"message,omitempty"`

	// NetworkInterfacesID
	NetworkInterfacesID []string `json:"networkInterfaces,omitempty"`

	// IP is the private ip address
	IP string `json:"ip,omitempty"`

	// ID of this PrivateEndpoint.
	ID string `json:"id,omitempty"`

	// Etag - A unique read-only string that changes whenever the resource is
	// updated.
	Etag string `json:"etag,omitempty"`

	// Type of this PrivateEndpoint.
	Type string `json:"type,omitempty"`
}

// +kubebuilder:object:root=true

// A PrivateEndpoint is a managed resource that represents an Azure Private Endpoint
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type PrivateEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrivateEndpointSpec   `json:"spec"`
	Status PrivateEndpointStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PrivateEndpointList contains a list of PrivateEndpoint items
type PrivateEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrivateEndpoint `json:"items"`
}
