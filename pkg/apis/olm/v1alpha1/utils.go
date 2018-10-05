package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSyndesisInstallPlan(namespace string) *InstallPlan {
	return &InstallPlan{
		TypeMeta: metav1.TypeMeta{
			APIVersion: groupName + "/" + version,
			Kind: "InstallPlan",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis.0.0.1-install",
			Namespace: namespace,
		},
		Spec: InstallPlanSpec{
			Approval: "Automatic",
			ClusterServiceVersionNames: []string{
				"syndesis-0.0.1",
			},
		},
	}
}
