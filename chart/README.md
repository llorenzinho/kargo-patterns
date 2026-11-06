# kargo-pattern Helm chart

Installs the `kargo-pattern` service with a ServiceAccount, Deployment, Service
and optional Ingress.

## Install

From the published OCI registry:

```sh
helm install kargo-pattern \
  oci://ghcr.io/llorenzinho/charts/kargo-pattern \
  --version 0.1.0 \
  --set image.tag=v0.1.0
```

From the repository:

```sh
helm install kargo-pattern ./chart \
  --set image.repository=ghcr.io/llorenzinho/kargo-pattern \
  --set image.tag=v0.1.0
```

## Values

| Key | Default | Description |
|---|---|---|
| `replicaCount` | `1` | Number of replicas |
| `image.repository` | `ghcr.io/llorenzinho/kargo-pattern` | Container image |
| `image.tag` | `Chart.AppVersion` | Image tag |
| `serviceAccount.create` | `true` | Create a dedicated ServiceAccount |
| `service.port` | `80` | Service port |
| `ingress.enabled` | `true` | Create the Ingress |
| `ingress.className` | `nginx` | Ingress class |
| `ingress.hosts[0].host` | `kargo-pattern.localhost` | Ingress host |
| `env.LOG_LEVEL` | `info` | Application log level |
