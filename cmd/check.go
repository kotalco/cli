package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkCmd checks cluster compliance
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks cluster compliance",
	Long:  "Checks the underlying cluster is suitable for installing Kotal components",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Check underlying cluster compliance")
		fmt.Println()

		var client client.Client
		var disclient *discovery.DiscoveryClient
		var err error

		if client, disclient, err = CanCreateKubernetesClient(); err != nil {
			fmt.Printf("❌ can create Kubernetes client: %s", err)
			return
		} else {
			fmt.Println("✔️ can create Kubernetes client")
		}

		if err = CanQueryKubernetesAPI(client); err != nil {
			fmt.Printf("❌ can query Kubernetes API: %s", err)
			return
		} else {
			fmt.Println("✔️ can query Kubernetes API")
		}

		if err = MinimumKubernetesVersion(disclient); err != nil {
			fmt.Printf("❌ is running minimum Kubernetes version: %s", err)
			return
		} else {
			fmt.Println("✔️ is running minimum Kubernetes version")
		}

		if err = NamespaceExists(client); err != nil {
			fmt.Printf("❌ kotal namespace doesn't exists: %s", err)
			return
		} else {
			fmt.Println("✔️ kotal namespace doesn't exist")
		}

		var ns string
		if ns, err = CanCreateNamespaces(client); err != nil {
			fmt.Printf("❌ can create Namespaces: %s", err)
			return
		} else {
			fmt.Println("✔️ can create Namespaces")
		}

		defer func() {
			key := types.NamespacedName{Name: ns}
			ns := corev1.Namespace{}
			client.Get(context.Background(), key, &ns)
			client.Delete(context.Background(), &ns)
		}()

		var role string
		if role, err = CanCreateClusterRoles(client); err != nil {
			fmt.Printf("❌ can create ClusterRoles: %s", err)
			return
		} else {
			fmt.Println("✔️ can create ClusterRoles")
		}

		var account string
		if account, err = CanCreateServiceAccounts(client); err != nil {
			fmt.Printf("❌ can create ServiceAccounts: %s", err)
			return
		} else {
			fmt.Println("✔️ can create ServiceAccounts")
		}

		if err = CanCreateClusterRoleBindings(client, role, account); err != nil {
			fmt.Printf("❌ can create ClusterRoleBindings: %s", err)
			return
		} else {
			fmt.Println("✔️ can create ClusterRoleBindings")
		}

		defer func() {
			key := types.NamespacedName{Name: role}
			role := rbacv1.ClusterRole{}
			client.Get(context.Background(), key, &role)
			client.Delete(context.Background(), &role)
		}()

		defer func() {
			key := types.NamespacedName{Name: account, Namespace: ns}
			sa := corev1.ServiceAccount{}
			client.Get(context.Background(), key, &sa)
			client.Delete(context.Background(), &sa)
		}()

		if err = CanCreateCustomResourceDefinitions(client); err != nil {
			fmt.Printf("❌ can create CustomResourceDefinitions: %s", err)
			return
		} else {
			fmt.Println("✔️ can create CustomResourceDefinitions")
		}

		go func() {
			key := types.NamespacedName{Name: "checks.cli.kotal.io"}
			crd := apiextensionsv1.CustomResourceDefinition{}
			client.Get(context.Background(), key, &crd)
			client.Delete(context.Background(), &crd)
		}()

		if err = CanCreateServices(client, ns); err != nil {
			fmt.Printf("❌ can create Services: %s", err)
			return
		} else {
			fmt.Println("✔️ can create Services")
		}

		if err = CanCreateDeployments(client, ns); err != nil {
			fmt.Printf("❌ can create Deployments: %s", err)
			return
		} else {
			fmt.Println("✔️ can create Deployments")
		}

		if err = CanCreateSecrets(client, ns); err != nil {
			fmt.Printf("❌ can create Secrets: %s", err)
			return
		} else {
			fmt.Println("✔️ can create Secrets")
		}

		if err = CanCreateMutatingWebhookConfiguration(client); err != nil {
			fmt.Printf("❌ can create MutatingWebhookConfiguration: %s", err)
			return
		} else {
			fmt.Println("✔️ can create MutatingWebhookConfiguration")
		}

		if err = CanCreateValidatingWebhookConfiguration(client); err != nil {
			fmt.Printf("❌ can create ValidatingWebhookConfiguration: %s", err)
			return
		} else {
			fmt.Println("✔️ can create ValidatingWebhookConfiguration")
		}

		if err = CertManagerIsInstalled(client); err != nil {
			fmt.Printf("❌ cert-manager is installed: %s", err)
			return
		} else {
			fmt.Println("✔️ cert-manager is installed")
		}

		if err = CanCreateCertManagerIssuer(client, ns); err != nil {
			fmt.Printf("❌ can create cert-manager Issuer: %s", err)
			return
		} else {
			fmt.Println("✔️ can create cert-manager Issuer")
		}

		// TODO: Can create cert-manager Certificates

	},
}

// CanCreateKubernetesClient checks if we can create Kubernetes client from config
func CanCreateKubernetesClient() (client.Client, *discovery.DiscoveryClient, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("error configuring Kubernetes API client: %w", err)
	}

	scheme := runtime.NewScheme()

	apiextensionsv1.AddToScheme(scheme)
	clientgoscheme.AddToScheme(scheme)
	cmv1.AddToScheme(scheme)

	opts := client.Options{Scheme: scheme}

	client, err := client.New(config, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating new Kubernetes client: %w", err)
	}

	disclient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return client, nil, fmt.Errorf("error creating new Kubernetes discovery client: %w", err)
	}

	return client, disclient, nil
}

// CanQueryKubernetesAPI checks if we can query Kubernetes API
func CanQueryKubernetesAPI(client client.Client) error {
	pods := corev1.PodList{}
	return client.List(context.Background(), &pods)
}

// NamespaceExists checks if kotal namespace exists
func NamespaceExists(client client.Client) error {
	key := types.NamespacedName{
		Name: "kotal",
	}
	ns := corev1.Namespace{}

	err := client.Get(context.Background(), key, &ns)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if !ns.CreationTimestamp.IsZero() {
		return fmt.Errorf("namespace does exist")
	}

	return fmt.Errorf("error getting namespace: %w", err)
}

func CanCreateNamespaces(client client.Client) (string, error) {
	id := uuid.NewString()

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
	}
	return id, client.Create(context.Background(), &ns)
}

func CanCreateClusterRoles(client client.Client) (string, error) {
	id := uuid.NewString()
	// dummy cluster role
	role := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
	}
	return id, client.Create(context.Background(), &role)

}

func CanCreateClusterRoleBindings(client client.Client, role, sa string) error {
	id := uuid.NewString()
	binding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa,
				Namespace: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     role,
		},
	}

	defer func() {
		key := types.NamespacedName{Name: role}
		binding := rbacv1.ClusterRoleBinding{}
		client.Get(context.Background(), key, &binding)
		client.Delete(context.Background(), &binding)
	}()

	return client.Create(context.Background(), &binding)
}

func CanCreateServiceAccounts(client client.Client) (string, error) {
	id := uuid.NewString()
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: "default",
		},
	}

	return id, client.Create(context.Background(), &sa)
}

func CanCreateCustomResourceDefinitions(client client.Client) error {
	crd := apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "checks.cli.kotal.io",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "cli.kotal.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "Check",
				ListKind: "CheckList",
				Singular: "check",
				Plural:   "checks",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  false,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"apiVersion": {
									Type: "string",
								},
								"kind": {
									Type: "string",
								},
								"spec": {
									Type: "object",
								},
							},
						},
					},
				},
			},
		},
	}

	return client.Create(context.Background(), &crd)

}

func CanCreateServices(client client.Client, ns string) error {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy",
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "dummy",
					Protocol:   corev1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.IntOrString{IntVal: 3000},
				},
			},
		},
	}
	return client.Create(context.Background(), &svc)
}

func CanCreateDeployments(client client.Client, ns string) error {
	id := uuid.NewString()
	sa := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "box",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "box",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "box",
							Image: "busybox",
						},
					},
				},
			},
		},
	}
	return client.Create(context.Background(), &sa)
}

func CanCreateSecrets(client client.Client, ns string) error {
	id := uuid.NewString()
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: ns,
		},
		StringData: map[string]string{
			"secret": "I am Satoshi",
		},
	}

	return client.Create(context.Background(), &secret)
}

func MinimumKubernetesVersion(client *discovery.DiscoveryClient) error {
	info, err := client.ServerVersion()
	if err != nil {
		return fmt.Errorf("error getting server info: %w", err)
	}

	minor, _ := strconv.Atoi(info.Minor)
	major, _ := strconv.Atoi(info.Major)

	if major < 1 || minor < 19 {
		return fmt.Errorf("cluster version is %s, minimum required version is v1.19", info)
	}

	return nil
}

func CanCreateMutatingWebhookConfiguration(client client.Client) error {
	id := uuid.NewString()
	hook := admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{},
	}

	defer func() {
		key := types.NamespacedName{Name: id}
		hook := admissionregistrationv1.MutatingWebhookConfiguration{}
		client.Get(context.Background(), key, &hook)
		client.Delete(context.Background(), &hook)
	}()

	return client.Create(context.Background(), &hook)
}

func CanCreateValidatingWebhookConfiguration(client client.Client) error {
	id := uuid.NewString()
	hook := admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{},
	}

	defer func() {
		key := types.NamespacedName{Name: id}
		hook := admissionregistrationv1.ValidatingWebhookConfiguration{}
		client.Get(context.Background(), key, &hook)
		client.Delete(context.Background(), &hook)
	}()

	return client.Create(context.Background(), &hook)
}

func CertManagerIsInstalled(client client.Client) error {
	certs := cmv1.CertificateList{}
	return client.List(context.Background(), &certs)
}

func CanCreateCertManagerIssuer(client client.Client, ns string) error {
	issuer := cmv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "self-signer-issuer",
			Namespace: ns,
		},
		Spec: cmv1.IssuerSpec{
			IssuerConfig: cmv1.IssuerConfig{
				SelfSigned: &cmv1.SelfSignedIssuer{},
			},
		},
	}
	return client.Create(context.Background(), &issuer)
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
