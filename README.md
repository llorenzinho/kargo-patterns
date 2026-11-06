# kargo-pattern

Showcase repository for a Kubernetes promotion pipeline built with
[Kargo](https://kargo.akuity.io/) and ArgoCD.

The GitHub repository is `llorenzinho/kargo-patterns` (plural); the service,
chart and local cluster keep the `kargo-pattern` project name.

The repository contains:

- `app/`: a small Go HTTP service with structured JSON logs
- `chart/`: a Helm chart for the service
- `gitops/kargo/`: scaffold for the Kargo promotion pattern
- `gitops/cluster/`: scaffold for the ArgoCD ApplicationSet bootstrap

Kargo and ArgoCD are intentionally not implemented yet.

## Service

The service exposes:

| Endpoint | Description |
|---|---|
| `GET /` | Hello response with service and version |
| `GET /healthz` | Liveness probe |
| `GET /readyz` | Readiness probe |
| `GET /version` | Service, version and commit |

Every log record is JSON and includes the service build metadata:

```json
{"time":"2026-11-06T12:00:00Z","level":"INFO","msg":"http request","service":"kargo-pattern","version":"0.1.0","commit":"abc1234","method":"GET","path":"/version","status":200}
```

Configuration:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |

## Requirements

- Docker Engine or Docker Desktop running locally
- [mise](https://mise.jdx.dev/)

`mise.toml` pins Go, kind, kubectl, Helm, kustomize, Task, pre-commit,
Docker CLI, golangci-lint, actionlint, ArgoCD CLI, Kargo CLI, yq and jq.

mise installs the Docker **CLI**, not the Docker daemon.

## Quick start

```sh
mise trust
task setup
task doctor
task app:run
```

Then call the service:

```sh
curl http://localhost:8080/version
curl http://localhost:8080/healthz
```

## Local Kubernetes demo

```sh
task e2e
```

The task creates a kind cluster with ingress ports `8080`/`8443`, builds and
loads the image, installs ingress-nginx, installs the chart and checks
`http://kargo-pattern.localhost:8080/version`.

Individual steps:

```sh
task cluster:create
task image:build
task image:load
task ingress:install
task chart:install
```

Remove the cluster with:

```sh
task cluster:delete
```

### Troubleshooting: kube-proxy crash loops

kind nodes share the host inotify limits. With the common default
`fs.inotify.max_user_instances=128`, `kube-proxy` crash-loops with
`fsnotify watcher init: too many open files`, no Service VIPs are programmed and
every pod fails to reach the API server. `task cluster:create` raises the limits
inside the privileged kind node; to change them host-wide instead:

```sh
sudo sysctl -w fs.inotify.max_user_instances=1024 fs.inotify.max_user_watches=524288
```

## Helm chart

The chart creates a ServiceAccount, Deployment, Service and optional Ingress.
The default image tag is `Chart.AppVersion`; pass `image.tag` explicitly when
the chart and the application versions are not aligned.

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
  --set image.repository=ghcr.io/llorenzinho/kargo-patterns \
  --set image.tag=v0.1.0
```

## CI/CD

- `ci.yml`: pre-commit, Go checks, Helm lint/template and image build
- `release-please.yml`: release PRs and tags for `app` and `chart`
- `image.yml`: builds and pushes `ghcr.io/<owner>/kargo-patterns`
- `chart-publish.yml`: packages the chart, pushes it to
  `oci://ghcr.io/<owner>/charts` and attaches the `.tgz` to the GitHub release

The app and chart are released independently by release-please:

| Path changed | Release tag | Artifact |
|---|---|---|
| `app/` | `vX.Y.Z` | `ghcr.io/<owner>/kargo-patterns:vX.Y.Z` |
| `chart/` | `chart-vX.Y.Z` | `oci://ghcr.io/<owner>/charts/kargo-pattern:X.Y.Z` |

Commits follow [Conventional Commits](https://www.conventionalcommits.org/).
The commit-msg hook validates them locally.

`GHCR` packages are private by default. After the first push, set the container
package visibility to **public** in the repository package settings.

### Optional release-please token

The default `GITHUB_TOKEN` cannot trigger other workflows from the release PR.
CI still runs on normal pull requests. To also run CI on release PRs, create a
fine-grained personal access token with `contents: write` and
`pull-requests: write` permissions and store it as the `RELEASE_PLEASE_TOKEN`
repository secret.

## GitOps roadmap

- `gitops/cluster/`: ArgoCD ApplicationSet bootstrap
- `gitops/kargo/`: Kargo Project, Warehouse, Stages and PromotionTasks

The Kargo Warehouse will watch the image repository, and each Stage will
promote the image tag through the GitOps desired state.
