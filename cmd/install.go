package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install kotal components",
	Long:  "Install kotal operator, backend, and dashboard.",
	Run: func(cmd *cobra.Command, args []string) {

		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			// TODO: default to latest version
			version = "0.1-alpha.6"
		}

		fmt.Println("🚀 Installing Kotal operator")

		url := fmt.Sprintf("https://github.com/kotalco/kotal/releases/download/v%s/kotal.yaml", version)
		c := exec.Command("kubectl", "apply", "-f", url)

		var outb, errb bytes.Buffer
		c.Stderr = &errb
		c.Stdout = &outb

		if err := c.Run(); err != nil {
			fmt.Printf("🥵 %s\n", errb.String())
			return
		}

		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			fmt.Printf("%s\n", outb.String())
		}

		fmt.Println("👍 Kotal operator has been installed")
		fmt.Println("⏰ Waiting for the operator to start successfully")

		c = exec.Command("kubectl", "wait", "-n", "kotal", "--for=condition=available", "deployments/controller-manager", "--timeout=600s")
		c.Stderr = &errb
		c.Stdout = &outb

		if err := c.Run(); err != nil {
			fmt.Printf("🥵 %s\n", errb.String())
			return
		}

		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			fmt.Printf("%s\n", outb.String())
		}

		fmt.Println("🙌 Operator is up and running")

		// TODO: install api server
		// TODO: install dashboard

	},
}

func init() {

	rootCmd.AddCommand(installCmd)

	installCmd.Flags().String("version", "", "kotal version")
	installCmd.Flags().Bool("verbose", false, "verbose output")

}
