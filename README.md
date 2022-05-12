# kubectl-credentials-helper

This is a `kubectl` credentials helper. It is using your local OS keychain to store sensitive content of your `KUBECONFIG`.

To install:

```bash
go install github.com/plumber-cd/kubectl-credentials-helper
```

You can also download this from releases.

To secure your existing `KUBECONFIG`, use:

```bash
kubectl-credentials-helper secure --kubeconfig path
```

If `--kubeconfig` was not provided - it will try to find `KUBECONFIG` env variable and as a last resort - default user home `~/.kube/config`.

This will save all sensitive info from a local `KUBECONFIG` to your OS specific keychain. Sensitive considered the following:

- `.user.username`
- `.user.password`
- `.user.client-certificate-data`
- `.user.client-key-data`

For user entries that it finds, it will replace in your `KUBECONFIG` with the following:

```yaml
apiVersion: v1
kind: Config
users:
- name: <name>
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1
      command: kubectl-credentials-helper
      provideClusterInfo: true
      interactiveMode: Never
```

Now, every time you are trying to access this cluster - the helper will fetch sensitive info from OS specific keychain.
