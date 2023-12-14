/* AUTO GENERATED CODE */
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// WaitingRoomSpecApplyConfiguration represents an declarative configuration of the WaitingRoomSpec type for use
// with apply.
type WaitingRoomSpecApplyConfiguration struct {
	Path           *string `json:"path,omitempty"`
	ActiveUsers    *int    `json:"activeUsers,omitempty"`
	Schema         *string `json:"schema,omitempty"`
	Host           *string `json:"host,omitempty"`
	BackendSvcAddr *string `json:"backendSvcAddr,omitempty"`
	BackendSvcPort *int    `json:"backendSvcPort,omitempty"`
}

// WaitingRoomSpecApplyConfiguration constructs an declarative configuration of the WaitingRoomSpec type for use with
// apply.
func WaitingRoomSpec() *WaitingRoomSpecApplyConfiguration {
	return &WaitingRoomSpecApplyConfiguration{}
}

// WithPath sets the Path field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Path field is set to the value of the last call.
func (b *WaitingRoomSpecApplyConfiguration) WithPath(value string) *WaitingRoomSpecApplyConfiguration {
	b.Path = &value
	return b
}

// WithActiveUsers sets the ActiveUsers field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ActiveUsers field is set to the value of the last call.
func (b *WaitingRoomSpecApplyConfiguration) WithActiveUsers(value int) *WaitingRoomSpecApplyConfiguration {
	b.ActiveUsers = &value
	return b
}

// WithSchema sets the Schema field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Schema field is set to the value of the last call.
func (b *WaitingRoomSpecApplyConfiguration) WithSchema(value string) *WaitingRoomSpecApplyConfiguration {
	b.Schema = &value
	return b
}

// WithHost sets the Host field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Host field is set to the value of the last call.
func (b *WaitingRoomSpecApplyConfiguration) WithHost(value string) *WaitingRoomSpecApplyConfiguration {
	b.Host = &value
	return b
}

// WithBackendSvcAddr sets the BackendSvcAddr field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BackendSvcAddr field is set to the value of the last call.
func (b *WaitingRoomSpecApplyConfiguration) WithBackendSvcAddr(value string) *WaitingRoomSpecApplyConfiguration {
	b.BackendSvcAddr = &value
	return b
}

// WithBackendSvcPort sets the BackendSvcPort field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BackendSvcPort field is set to the value of the last call.
func (b *WaitingRoomSpecApplyConfiguration) WithBackendSvcPort(value int) *WaitingRoomSpecApplyConfiguration {
	b.BackendSvcPort = &value
	return b
}