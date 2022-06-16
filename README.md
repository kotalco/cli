# Kotal CLI


## Build from source

```bash
git clone git@github.com:kotalco/cli.git
cd cli
go build -o kotal
./kotal --help
```

## check

To check that if we can safely install kotal components into the Kubernetes cluster we've access to, use `kotal check` command

```
./kotal check
```

It will return output similar to the following:

```
Check underlying cluster compliance:

✔️ can create Kubernetes client
✔️ can query Kubernetes API
✔️ is running minimum Kubernetes version
✔️ kotal namespace doesn't exist
✔️ can create Namespaces
✔️ can create ClusterRoles
✔️ can create ServiceAccounts
✔️ can create ClusterRoleBindings
✔️ can create CustomResourceDefinitions
✔️ can create Services
✔️ can create Deployments
✔️ can create Secrets
✔️ can create MutatingWebhookConfigurations
✔️ can create ValidatingWebhookConfigurations
✔️ cert-manager is installed
✔️ can create cert-manager Issuers
✔️ can create cert-manager Certificates

🔥 kotal can be installed
```

## install

If no issues have been reported using `kotal check` command, you can go ahead and install kotal components using `kotal install` command:

```bash
./kotal install --version 0.1-alpha.6
```

It will return output similar to the following:

```
🚀 Installing Kotal operator
👍 Kotal operator has been installed
⏰ Waiting for the operator to start successfully
🙌 Operator is up and running
```

## dashboard

TODO

## diagnose

TODO

## upgrade

TODO