package controller

import (
	wr "github.com/hamedetemaad/lineq-operator/pkg/waitingroom"
	wrv1alpha1 "github.com/hamedetemaad/lineq-operator/pkg/waitingroom/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createIngress(newWaitingRoom *wrv1alpha1.WaitingRoom, namespace string) *netv1.Ingress {
	return &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newWaitingRoom.ObjectMeta.Name,
			Namespace: namespace,
			Labels:    make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(
					newWaitingRoom,
					wrv1alpha1.SchemeGroupVersion.WithKind(wr.WaitingRoomKind),
				),
			},
		},
		Spec: createIngressSpec(newWaitingRoom.Name, namespace, newWaitingRoom.Spec.Path, newWaitingRoom.Spec.ActiveUsers, newWaitingRoom.Spec.Host, newWaitingRoom.Spec.Schema, newWaitingRoom.Spec.BackendSvcAddr, newWaitingRoom.Spec.BackendSvcPort),
	}
}

func createIngressSpec(name, namespace, path string, activeUsers int, host string, scheme string, backSvcAddr string, backSvcPort int) netv1.IngressSpec {
	pathTypeImplementationSpecific := netv1.PathTypeImplementationSpecific
	ingressClassName := "haproxy"
	ingressRule := netv1.IngressRule{
		Host: host,
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &pathTypeImplementationSpecific,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: backSvcAddr,
								Port: netv1.ServiceBackendPort{
									Number: int32(backSvcPort),
								},
							},
						},
					},
				},
			},
		},
	}

	ingressSpec := netv1.IngressSpec{
		IngressClassName: &ingressClassName,
		Rules:            []netv1.IngressRule{ingressRule},
	}
	return ingressSpec
}
