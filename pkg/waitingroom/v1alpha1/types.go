package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WaitingRoom struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec WaitingRoomSpec `json:"spec"`
}

type WaitingRoomSpec struct {
	Path           string `json:"path"`
	ActiveUsers    int    `json:"activeUsers"`
	Schema         string `json:"schema"`
	Host           string `json:"host"`
	BackendSvcAddr string `json:"backendSvcAddr"`
	BackendSvcPort int    `json:"backendSvcPort"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WaitingRoomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WaitingRoom `json:"items"`
}