package info

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KindIPPool     = "IPPool"
	KindIPPoolList = "IPPoolList"
)

type Mode string

const (
	Undefined   Mode = ""
	Always           = "always"
	CrossSubnet      = "cross-subnet"
)
const DefaultMode = Always

type IPIPConfiguration struct {
	Enabled bool `json:"enabled,omitempty"`
	Mode    Mode `json:"mode,omitempty" validate:"ipIpMode"`
}
type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IPPoolSpec `json:"spec,omitempty"`
}
type IPPoolSpec struct {
	CIDR          string             `json:"cidr" validate:"net"`
	VXLANMode     VXLANMode          `json:"vxlanMode,omitempty" validate:"omitempty,vxlanMode"`
	IPIPMode      IPIPMode           `json:"ipipMode,omitempty" validate:"omitempty,ipIpMode"`
	NATOutgoing   bool               `json:"natOutgoing,omitempty"`
	Disabled      bool               `json:"disabled,omitempty"`
	BlockSize     int                `json:"blockSize,omitempty"`
	NodeSelector  string             `json:"nodeSelector,omitempty" validate:"omitempty,selector"`
	IPIP          *IPIPConfiguration `json:"ipip,omitempty" validate:"omitempty,mustBeNil"`
	NATOutgoingV1 bool               `json:"nat-outgoing,omitempty" validate:"omitempty,mustBeFalse"`
}
type VXLANMode string
type IPIPMode string
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPPool `json:"items"`
}
