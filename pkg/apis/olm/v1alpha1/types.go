package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstallPlan struct {
	metav1.TypeMeta        `json:",inline"`
	metav1.ObjectMeta      `json:"metadata"`
	Spec InstallPlanSpec   `json:"spec"`
}

type InstallPlanSpec struct {
	metav1.TypeMeta                         `json:",inline"`
	metav1.ObjectMeta                       `json:"metadata"`
	Approval string                         `json:"approval"`
	ClusterServiceVersionNames []string     `json:"clusterServiceVersionNames"`
}