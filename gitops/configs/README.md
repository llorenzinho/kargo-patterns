# Cluster configuration

Installs ArgoCD, then lets ArgoCD install cert-manager, Kargo, the Kargo pipeline and
the `kargo-pattern` service in the `dev`, `staging` and `prod` namespaces.

```
argocd/      ArgoCD itself: upstream install.yaml pinned to a tag, plus local patches
bootstrap/   root Application (app-of-apps), the only manifest applied by hand
apps/        reconciled by root: everything else
demo/app-values/      application configuration: default.yaml, dev.yaml, staging.yaml, prod.yaml
```

`demo/app-values/` is not applied to the cluster. Kargo Warehouses watch these files and each
promotion renders the chart with them: `default.yaml` is promoted with the image and the
chart, while each `<stage>.yaml` is used only by its own Stage. See
[`../kargo/README.md`](../kargo/README.md).

## Flow

```
root (main: gitops/configs/apps)
├── argocd                 main: gitops/configs/argocd
├── cert-manager           chart jetstack/cert-manager
├── kargo                  chart ghcr.io/akuity/kargo-charts/kargo
├── kargo-project          main: gitops/kargo
├── AppProject kargo-pattern
└── ApplicationSet kargo-pattern
    ├── kargo-pattern-dev      deploy/app/dev      -> namespace dev
    ├── kargo-pattern-staging  deploy/app/staging  -> namespace staging
    └── kargo-pattern-prod     deploy/app/prod     -> namespace prod
```

Each stage Application deploys `manifests/` from its `deploy/app/<stage>` branch.
Kargo owns those branches; do not edit them by hand. Every Application auto-syncs,
because these branches only change through a Kargo promotion. See
[`../kargo/README.md`](../kargo/README.md).

## Running everything

`task up` from the repository root is the entry point: see the
[README](../../README.md#gitops). It runs these steps:

```sh
task cluster:create
task ingress:install
task argocd:install
task kargo:admin
task argocd:bootstrap
task kargo:wait
task kargo:git-credentials
task up:info
```

ArgoCD reads from GitHub, so `gitops/` must be on `origin/main` first.

Secrets stay out of Git:

- `kargo:admin` generates the Kargo admin password on the first run and stores it in
  the `kargo-admin-credentials` Secret.
- `kargo:git-credentials` stores `KARGO_GIT_USERNAME` and `KARGO_GIT_TOKEN` from `.env`
  in the `release-train` namespace.

`server.insecure` on ArgoCD and `api.tls.enabled: false` on Kargo exist because
ingress-nginx serves plain HTTP on kind. They are for local use only: put TLS in front
of both before reusing this configuration on a real cluster.

## Expected state before the first promotion

Kargo creates the `deploy/app/<stage>` branches on the first promotion to a Stage.
Until then the stage Application reports:

```
ComparisonError: unable to resolve 'deploy/app/dev' to a commit SHA: revision not found
```

This is expected.

## Upgrading

- ArgoCD: change the tag in `argocd/kustomization.yaml` and the `argocd` version in
  `mise.toml`, then merge to `main`. The self-managed `argocd` Application applies the
  change. It never prunes automatically, so resources dropped upstream must be removed
  by hand.
- Kargo: change `targetRevision` in `apps/kargo.yaml` and the Kargo CLI version in
  `mise.toml`, then merge to `main`.

## Validation

`task gitops:validate` renders every directory and runs as part of `task ci`.
