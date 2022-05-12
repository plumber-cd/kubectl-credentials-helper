package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/plumber-cd/kubectl-credentials-helper/keychain"
)

func init() {
	undoCmd.Flags().StringP("kubeconfig", "c", "", "Kubeconfig path")

	rootCmd.AddCommand(undoCmd)
}

var undoCmd = &cobra.Command{
	Use:   "undo --kubeconfig path",
	Short: "This makes your kubeconfig insecure!",
	Run: func(cmd *cobra.Command, args []string) {
		path := findKubeConfig(cmd)
		cfg := loadAndBackup(path)

		executable, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}

		for name, user := range cfg.AuthInfos {
			fmt.Printf("Found user: %s\n", name)

			if user.Exec == nil || !strings.HasSuffix(user.Exec.Command, executable) {
				fmt.Printf("Skip user %s: doesn't seem to be configured to use this helper\n", name)
				continue
			}

			fmt.Printf("Looking up contexts for user: %s\n", name)

			for contextName, context := range cfg.Contexts {
				if context.AuthInfo != name {
					continue
				}

				fmt.Printf("Found context: %s\n", contextName)

				fmt.Printf("Looking up clusters for context: %s\n", contextName)
				for clusterName, cluster := range cfg.Clusters {
					if clusterName != context.Cluster {
						continue
					}

					fmt.Printf("Found cluster %s (%s)\n", clusterName, cluster.Server)

					secretName, secretB64, err := keychain.GetSecret(cluster.Server)
					if err != nil {
						if err == keychain.ErrorItemNotFound {
							continue
						}
						log.Fatal(err)
					}

					secret, err := base64.StdEncoding.DecodeString(secretB64)
					if err != nil {
						log.Fatal(err)
					}

					secretConfig, err := clientcmd.Load(secret)
					if err != nil {
						log.Fatal(err)
					}

					user.Username = secretConfig.AuthInfos[secretName].Username
					user.Password = secretConfig.AuthInfos[secretName].Password
					user.ClientCertificateData = secretConfig.AuthInfos[secretName].ClientCertificateData
					user.ClientKeyData = secretConfig.AuthInfos[secretName].ClientKeyData
					user.Exec = nil
					fmt.Printf("Restored from secret: %s (%s)\n", clusterName, cluster.Server)
					break
				}
			}

			fmt.Printf("Unsecured user: %s\n", name)
		}

		if err := clientcmd.WriteToFile(cfg, path); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Wrote: %s\n", path)
	},
}
