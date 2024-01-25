// Code generated by protoc-json-shim. DO NOT EDIT.
package pbproxystate

import (
	protojson "google.golang.org/protobuf/encoding/protojson"
)

// MarshalJSON is a custom marshaler for TrafficPermissions
func (this *TrafficPermissions) MarshalJSON() ([]byte, error) {
	str, err := TrafficPermissionsMarshaler.Marshal(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TrafficPermissions
func (this *TrafficPermissions) UnmarshalJSON(b []byte) error {
	return TrafficPermissionsUnmarshaler.Unmarshal(b, this)
}

// MarshalJSON is a custom marshaler for Permission
func (this *Permission) MarshalJSON() ([]byte, error) {
	str, err := TrafficPermissionsMarshaler.Marshal(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Permission
func (this *Permission) UnmarshalJSON(b []byte) error {
	return TrafficPermissionsUnmarshaler.Unmarshal(b, this)
}

// MarshalJSON is a custom marshaler for Principal
func (this *Principal) MarshalJSON() ([]byte, error) {
	str, err := TrafficPermissionsMarshaler.Marshal(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Principal
func (this *Principal) UnmarshalJSON(b []byte) error {
	return TrafficPermissionsUnmarshaler.Unmarshal(b, this)
}

// MarshalJSON is a custom marshaler for Spiffe
func (this *Spiffe) MarshalJSON() ([]byte, error) {
	str, err := TrafficPermissionsMarshaler.Marshal(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Spiffe
func (this *Spiffe) UnmarshalJSON(b []byte) error {
	return TrafficPermissionsUnmarshaler.Unmarshal(b, this)
}

// MarshalJSON is a custom marshaler for DestinationRule
func (this *DestinationRule) MarshalJSON() ([]byte, error) {
	str, err := TrafficPermissionsMarshaler.Marshal(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for DestinationRule
func (this *DestinationRule) UnmarshalJSON(b []byte) error {
	return TrafficPermissionsUnmarshaler.Unmarshal(b, this)
}

// MarshalJSON is a custom marshaler for DestinationRuleHeader
func (this *DestinationRuleHeader) MarshalJSON() ([]byte, error) {
	str, err := TrafficPermissionsMarshaler.Marshal(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for DestinationRuleHeader
func (this *DestinationRuleHeader) UnmarshalJSON(b []byte) error {
	return TrafficPermissionsUnmarshaler.Unmarshal(b, this)
}

var (
	TrafficPermissionsMarshaler   = &protojson.MarshalOptions{}
	TrafficPermissionsUnmarshaler = &protojson.UnmarshalOptions{DiscardUnknown: false}
)
