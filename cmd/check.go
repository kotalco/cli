package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// checkCmd checks cluster compliance
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks cluster compliance",
	Long:  "Checks the underlying cluster is suitable for installing Kotal components",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Check underlying cluster compliance")
		// Can create kubernetes client
		// Can query Kubernetes API
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

func init() {
	rootCmd.AddCommand(checkCmd)
}
