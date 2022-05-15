package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	clientauthentication "k8s.io/client-go/pkg/apis/clientauthentication/v1"
	"k8s.io/client-go/tools/clientcmd"
	cmdApi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/plumber-cd/kubectl-credentials-helper/keychain"
)

// OsFs is an instance of afero.NewOsFs
var OsFs = afero.Afero{Fs: afero.NewOsFs()}

// FileExists afero for some reason does not have such a function, so...
func FileExists(filename string) (bool, error) {
	e, err := OsFs.Exists(filename)
	if err != nil {
		return e, err
	}

	e, err = OsFs.IsDir(filename)
	if err != nil {
		return e, err
	}

	return !e, nil
}

func init() {
	secureCmd.Flags().StringP("kubeconfig", "c", "", "Kubeconfig path")
	secureCmd.Flags().StringP("user", "u", "", "Secure specific user instead of all")

	rootCmd.AddCommand(secureCmd)
}

var secureCmd = &cobra.Command{
	Use:   "secure --kubeconfig path",
	Short: "This makes your kubeconfig secure!",
	Run: func(cmd *cobra.Command, args []string) {
		path := findKubeConfig(cmd)
		cfg := loadAndBackup(path)

		executable, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}

		specificUser, err := cmd.Flags().GetString("user")
		if err != nil {
			log.Fatal(err)
		}
		if specificUser != "" {
			fmt.Printf("Looking up for a specific user %s\n", specificUser)
		}

		for name, user := range cfg.AuthInfos {
			if specificUser != "" && specificUser != name {
				fmt.Printf("Skip user %s: not a %s\n", name, specificUser)
				continue
			}

			if len(user.ClientCertificateData) == 0 && len(user.ClientKeyData) == 0 && user.Username == "" && user.Password == "" {
				fmt.Printf("Skip user %s: nothing to secure\n", name)
				continue
			}

			fmt.Printf("Found user: %s\n", name)

			newUser := cmdApi.AuthInfo{
				ClientCertificateData: user.ClientCertificateData,
				ClientKeyData:         user.ClientKeyData,
				Username:              user.Username,
				Password:              user.Password,
			}
			newCfg := cmdApi.Config{
				APIVersion: cfg.APIVersion,
				Kind:       cfg.Kind,
				AuthInfos: map[string]*cmdApi.AuthInfo{
					name: &newUser,
				},
			}
			newCfgJsonBytes, err := clientcmd.Write(newCfg)
			if err != nil {
				log.Fatal(err)
			}
			newCfgJsonB64 := base64.StdEncoding.EncodeToString(newCfgJsonBytes)

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

					var answer string
					fmt.Printf("Create secret for %s (%s)? Type 'yes': ", clusterName, cluster.Server)
					fmt.Scanf("%s", &answer)
					if answer != "yes" {
						fmt.Printf("Ok, skip %s\n", clusterName)
						continue
					}

					if err := keychain.CreateSecret(clusterName, cluster.Server, newCfgJsonB64); err != nil {
						if err == keychain.ErrorDuplicateItem {
							var answer string
							fmt.Printf("Secret %s already found in your keychain, replace? Type 'yes': ", clusterName)
							fmt.Scanf("%s", &answer)
							if answer != "yes" {
								fmt.Printf("Ok, skip %s\n", clusterName)
								continue
							}

							if err := keychain.DeleteSecret(cluster.Server); err != nil {
								log.Fatal(err)
							}
							if err := keychain.CreateSecret(clusterName, cluster.Server, newCfgJsonB64); err != nil {
								log.Fatal(err)
							}
							fmt.Printf("Secret %s replaced\n", clusterName)
						} else {
							log.Fatal(err)
						}
					}
					fmt.Printf("Created secret: %s (%s)\n", clusterName, cluster.Server)
				}
			}

			var answer string
			fmt.Printf("Remove sensitive parts from the user %s? Type 'yes': ", name)
			fmt.Scanf("%s", &answer)
			if answer != "yes" {
				fmt.Printf("Ok, skip %s\n", name)
				continue
			}
			user.ClientCertificateData = nil
			user.ClientKeyData = nil
			user.Username = ""
			user.Password = ""

			user.Exec = &cmdApi.ExecConfig{
				APIVersion:         clientauthentication.SchemeGroupVersion.String(),
				Command:            executable,
				ProvideClusterInfo: true,
				InteractiveMode:    cmdApi.NeverExecInteractiveMode,
			}

			fmt.Printf("Secured user: %s\n", name)
		}

		if err := clientcmd.WriteToFile(cfg, path); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Wrote: %s\n", path)
	},
}

func findKubeConfig(cmd *cobra.Command) string {
	path, err := cmd.Flags().GetString("kubeconfig")
	if err != nil {
		log.Fatal(err)
	}
	if path == "" {
		fmt.Printf("--kubeconfig was not set, trying KUBECONFIG\n")
		if val, ok := os.LookupEnv("KUBECONFIG"); ok {
			path = val
		}
	}
	if path == "" {
		fmt.Printf("KUBECONFIG was not set, trying user home default\n")
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(home, ".kube", "config")
		if ok, err := FileExists(path); !ok || err != nil {
			log.Fatalf("Have to either specify --kubeconfig or KUBECONFIG; default %s did not existed", path)
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	if ok, err := FileExists(abs); !ok || err != nil {
		log.Fatalf("%s did not existed", abs)
	}
	fmt.Printf("Found: %s\n", abs)

	return abs
}

func loadAndBackup(path string) cmdApi.Config {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: path},
		&clientcmd.ConfigOverrides{},
	).RawConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Read: %s\n", path)

	back := path + ".back"
	if err := clientcmd.WriteToFile(cfg, back); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Backup: %s\n", back)

	return cfg
}
