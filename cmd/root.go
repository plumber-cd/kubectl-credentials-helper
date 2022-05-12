package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/plumber-cd/kubectl-credentials-helper/keychain"
	"github.com/spf13/cobra"
	clientauthentication "k8s.io/client-go/pkg/apis/clientauthentication/v1"
	"k8s.io/client-go/tools/auth/exec"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kubectl-credentials-helper",
		Short: "It helps",
		Run: func(cmd *cobra.Command, args []string) {
			printfln("KUBERNETES_EXEC_INFO: %q", os.Getenv("KUBERNETES_EXEC_INFO"))

			ec, _, err := exec.LoadExecCredentialFromEnv()
			if err != nil {
				dief("load: %q", err.Error())
			}

			ecv1, ok := ec.(*clientauthentication.ExecCredential)
			if !ok {
				dief("cast failed: %#v\n", ec)
			}

			clusterEndpoint := ecv1.Spec.Cluster.Server
			if clusterEndpoint == "" {
				dief("empty cluster endpoint in the input")
			}
			printfln("cluster endpoint: %s", clusterEndpoint)

			clusterName, secretB64, err := keychain.GetSecret(clusterEndpoint)
			if err != nil {
				dief("secret: %s", err)
			}
			printfln("found secret %s", clusterName)

			secret, err := base64.StdEncoding.DecodeString(secretB64)
			if err != nil {
				dief("b64: %s", err)
			}

			cfg, err := clientcmd.Load(secret)
			if err != nil {
				dief("config: %s", err)
			}

			ecv1.APIVersion = clientauthentication.SchemeGroupVersion.String()
			ecv1.Status = &clientauthentication.ExecCredentialStatus{
				ClientCertificateData: string(cfg.AuthInfos[clusterName].ClientCertificateData),
				ClientKeyData:         string(cfg.AuthInfos[clusterName].ClientKeyData),
			}

			data, err := json.Marshal(ecv1)
			if err != nil {
				dief("marshal: %q", err.Error())
			}
			printfln("marshal: %q", string(data))

			fmt.Println(string(data))
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
}

func printfln(format string, a ...interface{}) {
	if os.Getenv("KUBECTL_CREDENTIALS_HELPER_DEBUG") == "true" {
		reallyPrintf(format+"\n", a...)
	}
}

func reallyPrintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "kubectl-credentials-helper> "+format, a...)
}

func dief(format string, a ...interface{}) {
	reallyPrintf("error: "+format+"\n", a...)
	os.Exit(1)
}
