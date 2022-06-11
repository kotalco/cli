package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
			fmt.Printf("❌ can't create Kubernetes client: %s", err)
			return
		} else {
			fmt.Println("✔️ can create Kubernetes client")
		}

		if err = CanQueryKubernetesAPI(client); err != nil {
			fmt.Printf("❌ can't query Kubernetes API: %s", err)
			return
		} else {
			fmt.Println("✔️ can query Kubernetes API")
		}

		// kotal namespace doesn't exist
		// Can create Namespaces
		// Can create ClusterRoles
		// Can create ClusterRoleBindings
		// Can create CustomResourceDefinitions
		// can create ServiceAccounts
		// Can create Services
		// Can create Deployments
		// Can create Secrets
		// Certificate manager is installed
		// Can create cert-manager Certificates
		// Can create cert-manager Issuers
		// Can create MutatingWebhookConfiguration
		// Can create ValidatingWebhookConfiguration

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

func init() {
	rootCmd.AddCommand(checkCmd)
}
