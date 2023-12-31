/* AUTO GENERATED CODE */
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfiguration

import (
	v1alpha1 "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1"
	waitingroomv1alpha1 "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1/apis/applyconfiguration/waitingroom/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=lineq.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithKind("WaitingRoom"):
		return &waitingroomv1alpha1.WaitingRoomApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("WaitingRoomSpec"):
		return &waitingroomv1alpha1.WaitingRoomSpecApplyConfiguration{}

	}
	return nil
}
