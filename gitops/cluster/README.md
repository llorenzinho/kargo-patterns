# Cluster bootstrap

Installs ArgoCD and deploys the `kargo-pattern` service to the `dev`, `staging`
and `prod` namespaces with an `ApplicationSet`.

```
argocd/      ArgoCD itself: upstream install.yaml pinned to a tag, plus local patches
bootstrap/   root Application (app-of-apps), the only manifest applied by hand
apps/        reconciled by root: self-managed ArgoCD, AppProject, ApplicationSet
```

## Flow

```
root (main: gitops/cluster/apps)
├── argocd                 main: gitops/cluster/argocd
├── AppProject kargo-pattern
└── ApplicationSet kargo-pattern
    ├── kargo-pattern-dev      deploy/app/dev      -> namespace dev
    ├── kargo-pattern-staging  deploy/app/staging  -> namespace staging
    └── kargo-pattern-prod     deploy/app/prod     -> namespace prod
```

Each stage Application deploys the rendered manifests at the root of its
`deploy/app/<stage>` branch. Kargo owns those branches; do not edit them by hand.
Every Application auto-syncs, because these branches only change through a Kargo
promotion. The production gate belongs to the Kargo Stage.

## Bootstrap

ArgoCD reads from GitHub, so `gitops/cluster/` must be on `origin/main` before
`argocd:bootstrap`.

```sh
task cluster:create
task ingress:install
task argocd:install
task argocd:bootstrap
```

The UI is at <http://argocd.localhost:8080>. Log in as `admin` with the password
from `task argocd:password`.

`server.insecure` is enabled because ingress-nginx serves plain HTTP on kind. It
is for local use only: put TLS in front of ArgoCD before reusing this
configuration on a real cluster.

## Expected state before Kargo exists

The `deploy/app/<stage>` branches are created by Kargo. Until then the three stage
Applications report:

```
ComparisonError: unable to resolve 'deploy/app/dev' to a commit SHA: revision not found
```

This is expected. They sync on their own once Kargo pushes the branches.

## Upgrading ArgoCD

Change the tag in `argocd/kustomization.yaml` and the `argocd` version in
`mise.toml`, then merge to `main`. The self-managed `argocd` Application applies
the change. It never prunes automatically, so resources dropped upstream must be
removed by hand.

## Validation

`task gitops:validate` renders every directory and runs as part of `task ci`.
