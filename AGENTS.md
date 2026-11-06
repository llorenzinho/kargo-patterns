# AGENTS.md

Compact notes for coding agents. See `README.md` for the full user-facing setup.
Only repo-specific, easy-to-get-wrong facts belong here.

## Layout and ownership

| Path | What it is |
|---|---|
| `app/` | Go module `github.com/llorenzinho/kargo-pattern/app`. Standard library only. |
| `chart/` | Helm chart, published as an OCI artifact. |
| `gitops/configs/` | ArgoCD install, app-of-apps root, Kargo install, stage `ApplicationSet`, app configuration in `demo/app-values/` (watched by Warehouses, not applied). |
| `gitops/kargo/` | Kargo pipeline only (Project `release-train`). App configuration does not belong here. |
| `hack/kind-config.yaml` | Local kind cluster definition. |
| `.github/workflows/` | CI and release pipelines. |

- The Go module root is `app/`, not the repository root. `go test ./...` from the
  root does not work; run Go commands inside `app/` or use `task app:*`.
- `app/` and `chart/` are owned by release-please. `gitops/` is deliberately
  outside both packages, so GitOps-only commits must not bump app/chart versions.
- The GitHub repo and image use the plural `kargo-patterns`; the service, chart,
  namespace, kind cluster and Go module keep the singular `kargo-pattern` name.
  Do not "fix" the module path or chart name to match the repo.

## Commands

- Bootstrap once: `mise trust && task setup` (pins tools, creates `.env`,
  installs pre-commit hooks).
- Run everything CI runs: `mise exec -- task ci`.
- Before finishing: `git add -A && mise exec -- pre-commit run --all-files`.
  Pre-commit only sees tracked/staged files, so new untracked files are skipped.
- Single test: `cd app && go test -race ./internal/httpapi -run TestEndpoints -v`.
- Run the service: `task app:run` (port 8080; `PORT` and `LOG_LEVEL`).
- Chart only on kind: `task e2e`. Whole solution with ArgoCD and Kargo: `task up`.

Prefer `mise exec -- <cmd>` (or activate mise) for repo tooling. If
`golangci-lint run` fails with a v1/v2 config-version error, a system v1 is
shadowing mise's pinned v2 and `.golangci.yml` uses the v2 schema. `task
app:lint` and the pre-commit hook already call the mise-pinned binary.

## Versions and releases

- Do not hand-edit versions: `app/internal/version/version.go` carries
  `// x-release-please-version`, and `chart/Chart.yaml` is managed by the
  `helm` release type. Update `.release-please-manifest.json` only when seeding
  a version.
- Two independent packages: `app` → tag `vX.Y.Z`; `chart` → tag `chart-vX.Y.Z`.
- Commit messages must be Conventional Commits (commit-msg hook). Release notes
  and version bumps derive from them.
- Image and chart are published only for releases (`v*` / `chart-v*` tags), never
  on `main` pushes. `ci.yml` still builds the image with `push: false`.
- `release-please.yml` calls the reusable `image.yml` / `chart-publish.yml`
  workflows directly. Do not remove that path and expect tag pushes to publish:
  tags created with the default `GITHUB_TOKEN` do not trigger workflows.
- `chart-publish.yml` asserts that `Chart.yaml` `version` equals the release
  version.
- Image: `ghcr.io/llorenzinho/kargo-patterns`; chart OCI:
  `oci://ghcr.io/llorenzinho/charts`. Images are linux/amd64, distroless and
  nonroot; build args `VERSION`/`COMMIT` populate `/version` and every log line.

## Helm chart

- `image.tag` defaults to `Chart.AppVersion`, but app and chart release
  independently. Always set `image.tag` explicitly (the Taskfile uses `dev`).
- Override ingress hosts with `--set-json 'ingress.hosts=[{...}]'`. Using
  `--set ingress.hosts[0].host=...` replaces the list element and drops `paths`,
  failing with `spec.rules[0].http.paths: Required value`.

## Local Kubernetes

- `task e2e` = kind cluster `kargo-pattern` → build/load image → install
  ingress-nginx → install the chart → curl
  `http://kargo-pattern.localhost:8080`.
- `hack/kind-config.yaml` maps host 8080→30080 and 8443→30443; host port 8080
  must be free.
- `task e2e` requires a running Docker daemon; mise installs only the CLI.
- `ingress:install` is separate from `chart:install`.
- If pods cannot reach `10.96.0.1:443` and `kube-proxy` crash-loops with
  `fsnotify watcher init: too many open files`, host inotify limits are too low.
  `task cluster:tune` raises them inside the privileged kind node; the
  host-wide alternative is in the README.

## ArgoCD (`gitops/configs/`)

- `task argocd:install` applies `gitops/configs/argocd` with server-side apply;
  the ApplicationSet CRD fails with client-side apply.
- ArgoCD and Kargo read from GitHub, not the working tree. `root` and the
  self-managed `argocd` Application fail with `app path does not exist` until
  `gitops/configs/` is on `origin/main`. `task up` checks this first.
- The ApplicationSet deploys `manifests/` from `deploy/app/<stage>` to namespaces
  `dev`, `staging`, `prod`. Kargo owns these branches. Until the first promotion,
  `ComparisonError ... revision not found` is the expected state.
- Pinned versions must match `mise.toml`: ArgoCD in
  `gitops/configs/argocd/kustomization.yaml`, Kargo in `gitops/configs/apps/kargo.yaml`.
- `server.insecure: "true"` on ArgoCD and `api.tls.enabled: false` on Kargo are
  kind-only settings for the HTTP ingress.
- Kargo's Ingress lives in `gitops/configs/apps/kargo-ingress.yaml`, not in the chart:
  `api.host` carries the port, and the chart would copy it into an invalid Ingress host.

## Kargo (`gitops/kargo/`)

- The Project is `release-train`, not `kargo-pattern`: `task e2e` already uses the
  `kargo-pattern` namespace, and Kargo adopts only namespaces labeled for it.
- Gates (`staging-gate`, `prod-gate`) are control flow Stages: no `promotionTemplate`
  and no `ProjectConfig` policy. Adding promotion steps would turn them into regular
  Stages and break the `MatchUpstream` trick below.
- `promotionPolicies` is per Stage, so `staging` and `prod` have auto-promotion enabled
  for their `values-<stage>` Freight. Their `app` request stays manual only because of
  `selectionPolicy: MatchUpstream` with a gate upstream. Do not remove it.
- Every step in `promotion-task.yaml` has an `if` guard on both Freight requests
  because Kargo does not retry a failed auto-promotion: an empty request would stall
  the Stage. Kargo forbids `if` on a `task` step, so the guard cannot move to
  `stages.yaml`. Keep `success()` in it: an explicit `if` replaces the default of
  skipping a step after a failure.
- The promotion artifact is the `vX.Y.Z` image tag published by `image.yml`. Never
  promote `latest`.
- Stage values must set a unique ingress host. The chart default is the same in all
  three namespaces and ingress-nginx rejects the duplicate.
- Render from the published OCI chart. Do not copy chart templates into `gitops/`.
- `kargo.akuity.io/authorized-stage` belongs in the `ApplicationSet` template; the
  controller overwrites the generated Applications.
- Never commit secrets. Admin and Git credentials are created by `task kargo:admin`
  and `task kargo:git-credentials`.
- Keep changes out of `app/` and `chart/` unless the application itself changes,
  so release-please does not cut unrelated releases.
