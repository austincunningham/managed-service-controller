package stub

import (
	"context"
	"github.com/integr8ly/integration-controller/pkg/errors"
	apis "github.com/integr8ly/managed-services-controller/pkg/apis/integreatly/v1alpha1"
	olm "github.com/integr8ly/managed-services-controller/pkg/apis/olm/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func createNamespace(namespace string) error{
	_, ns := k8sclient.GetKubeClient().Core().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})
	return ns
}

func createSyndesisOperator(namespace string) error {
	return sdk.Create(olm.NewSyndesisInstallPlan(namespace))
}

func newIntegrationControllerDeployment(namespace string) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-controller",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "integration-controller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "integration-controller",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "integration-controller",
					Containers: []corev1.Container{
						{
							Name: "integration-controller",
							// TODO: Add to config
							Image: "quay.io/integreatly/integration-controller:dev",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6000,
									Name: "metrics",
								},
							},
							Command: []string{
								"integration-controller",
								"--allow-insecure=true",
								"--log-level=debug",
							},
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name: "WATCH_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "OPERATOR_NAME",
									Value: "integration-controller",
								},
							},
						},
					},
				},
			},
		},
	}
}

func newServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-controller",
		},
	}
}

func newRole(namespace string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-controller",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"integreatly.org",
				},
				Resources: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
					"services",
					"endpoints",
					"persistentvolumeclaims",
					"events",
					"configmaps",
					"secrets",
				},
				Verbs: []string{
					"*",
				},
			},
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"deployments",
					"daemonsets",
					"replicasets",
					"statefulsets",
				},
				Verbs: []string{
					"*",
				},
			},
			{
				APIGroups: []string{
					"syndesis.io",
				},
				Resources: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
		},
	}
}

func newRoleBinding() *rbacv1.RoleBinding{
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-integration-controller",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: "integration-controller",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "integration-controller",
			},
		},
	}
}

func enmasseConfigMapRoleBinding(namespace string) *rbacv1.RoleBinding{
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "integration-controller-enmasse-view-",
			Namespace: "enmasse",
			Labels: map[string]string{
				"for": "integration-controller",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "enmasse-integration-viewer",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "integration-controller",
				Namespace: namespace,
			},
		},
	}
}


func routesAndServicesRoleBinding(namespace string) *rbacv1.RoleBinding{
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "integration-controller-route-viewer-",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "route-service-viewer",
		},
		//TODO: extract any reusable objects
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "integration-controller",
				Namespace: namespace,
			},
		},
	}
}


func createIntegrationController(namespace string) error {
	k8sc := k8sclient.GetKubeClient()
	logrus.Info("Creating controller role")
	r := newRole(namespace)
	_, err := k8sc.Rbac().Roles(namespace).Create(r);if err != nil {
		return err
	}
	logrus.Info("Creating service account")
	s := newServiceAccount()
	_, err = k8sc.Core().ServiceAccounts(namespace).Create(s);if err != nil {
		return err
	}
	logrus.Info("Creating controller role binding")
	rb := newRoleBinding()
	_, err = k8sc.Rbac().RoleBindings(namespace).Create(rb);if err != nil {
		return err
	}
	logrus.Info("Creating enmasse rolebinding")
	erb := enmasseConfigMapRoleBinding(namespace)
	// TODO: add to config
	_, err = k8sc.Rbac().RoleBindings("enmasse").Create(erb);if err != nil {
		return err
	}
	logrus.Info("Creating routes and services rolebinding")
	rsrb := routesAndServicesRoleBinding(namespace)
	_, err = k8sc.Rbac().RoleBindings(namespace).Create(rsrb);if err != nil {
		return err
	}
	logrus.Info("deploying the controller")
	d := newIntegrationControllerDeployment(namespace);
	_, err = k8sc.Apps().Deployments(namespace).Create(d);if err != nil {
		return err
	}

	return nil
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *apis.ManagedServiceNamespace:
		logrus.Info("ManagedServiceNamespace event detected ", o.Name)

		ns := o.Spec.ManagedNamespace
		if event.Deleted == true {
			logrus.Info("Deleting ManagedServiceNamespace: ", o.Name)
			err := k8sclient.GetKubeClient().Core().Namespaces().Delete(ns, &metav1.DeleteOptions{})
			if errors.IsAlreadyExistsErr(err) == true {
				logrus.Info("ManagedServiceNamespace already deleted. Ignoring", o.Name)
				return nil
			}
			return err
		} else {
			err := createNamespace(ns)
			if err != nil {
				if errors.IsAlreadyExistsErr(err) == true {
					logrus.Info("ManagedServiceNamespace already exists. Ignoring", o.Name)
					return nil
				}
				return err
			}

			logrus.Info("Created ManagedServiceNamespace: ", o.Name)
			logrus.Info("Setting up ManagedServiceNamespace: ", o.Name)

			logrus.Info("Creating Syndesis operator")
			err = createSyndesisOperator(ns);if err != nil {
				return err
			}

			logrus.Info("Creating Integration Controller ")
			err = createIntegrationController(ns);if err != nil {
				return err
			}
		}
	}
	return nil
}
