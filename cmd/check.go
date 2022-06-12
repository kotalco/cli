package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
		var err error

		if client, err = CanCreateKubernetesClient(); err != nil {
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

		if err = CanCreateClusterRoles(client); err != nil {
			fmt.Printf("❌ can create ClusterRoles: %s", err)
			return
		} else {
			fmt.Println("✔️ can create ClusterRoles")
		}

		// TODO: Can create ClusterRoleBindings
		// TODO: Can create CustomResourceDefinitions
		// TODO: can create ServiceAccounts
		// TODO: Can create Services
		// TODO: Can create Deployments
		// TODO: Can create Secrets
		// TODO: Certificate manager is installed
		// TODO: Can create cert-manager Certificates
		// TODO: Can create cert-manager Issuers
		// TODO: Can create MutatingWebhookConfiguration
		// TODO: Can create ValidatingWebhookConfiguration

	},
}

// CanCreateKubernetesClient checks if we can create Kubernetes client from config
func CanCreateKubernetesClient() (client.Client, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error configuring Kubernetes API client: %w", err)
	}

	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)

	opts := client.Options{Scheme: scheme}

	client, err := client.New(config, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating new Kubernetes client: %w", err)
	}

	return client, nil
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

func CanCreateClusterRoles(client client.Client) error {
	id := uuid.NewString()
	// dummy cluster role
	role := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
	}
	defer func() {
		key := types.NamespacedName{Name: id}
		role := rbacv1.ClusterRole{}
		client.Get(context.Background(), key, &role)
		client.Delete(context.Background(), &role)
	}()
	return client.Create(context.Background(), &role)

}

func init() {
	rootCmd.AddCommand(checkCmd)
}
