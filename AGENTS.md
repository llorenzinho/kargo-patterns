# AGENTS.md

Compact notes for coding agents. See `README.md` for the full user-facing setup.
Only repo-specific, easy-to-get-wrong facts belong here.

## Layout and ownership

| Path | What it is |
|---|---|
| `app/` | Go module `github.com/llorenzinho/kargo-pattern/app`. Standard library only. |
| `chart/` | Helm chart, published as an OCI artifact. |
| `gitops/cluster/` | ArgoCD install, app-of-apps root and stage `ApplicationSet`. |
| `gitops/kargo/` | Scaffold only. No Kargo manifests exist yet. |
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
- Full local stack: `task e2e`.

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

## ArgoCD (`gitops/cluster/`)

- `task argocd:install` applies `gitops/cluster/argocd` with server-side apply;
  the ApplicationSet CRD fails with client-side apply. `task e2e` does not install
  ArgoCD.
- ArgoCD reads from GitHub, not the working tree. `root` and the self-managed
  `argocd` Application fail with `app path does not exist` until `gitops/cluster/`
  is on `origin/main`.
- The ApplicationSet deploys `deploy/app/<stage>` (rendered manifests at the branch
  root) to namespaces `dev`, `staging`, `prod`. Kargo owns these branches. Until it
  pushes them, `ComparisonError ... revision not found` is the expected state.
- The ArgoCD version is pinned in `gitops/cluster/argocd/kustomization.yaml` and
  must match `argocd` in `mise.toml`.
- `server.insecure: "true"` is a kind-only patch for the HTTP ingress.

## Kargo scaffold (to be implemented)

`gitops/kargo/` currently contains only a README. Kargo is intentionally not
installed by any task, including `task e2e`.

If you implement this part:

- Put all promotion resources in `gitops/kargo/`:
  - a Kargo `Project` for the service,
  - a `Warehouse` subscribing to `ghcr.io/llorenzinho/kargo-patterns`,
  - `Stage`s (for example `dev`, `staging`, `prod`) whose promotion steps update
    this repository's desired state instead of applying directly,
  - `PromotionTask`s and verification steps; keep production gated.
- Contract with the rest of the repo:
  - The promotion artifact is the image tag published by `image.yml`
    (`vX.Y.Z`, plus `sha-*`). Never promote `latest`.
  - Rendering must set `image.tag` explicitly per stage, because
    `Chart.AppVersion` is not the app image tag.
  - Rendering must set a unique ingress host per stage. The chart default is the
    same in all three namespaces and ingress-nginx rejects the duplicate.
  - Render from `chart/` (or the published OCI chart) and push to
    `deploy/app/<stage>`. Do not copy chart templates into `gitops/`.
  - Add `kargo.akuity.io/authorized-stage` to the `ApplicationSet` template, or
    the controller overwrites it on the generated Applications.
  - Never commit secrets (GHCR pull credentials, kubeconfigs); reference
    SealedSecrets/ExternalSecrets and document placeholders.
- Keep changes out of `app/` and `chart/` unless the application itself changes,
  so release-please does not cut unrelated releases.
- Validation: extend `task gitops:validate`, which already runs in `task ci`, to
  render `gitops/kargo/` with `kustomize build`. Document local bootstrap commands
  in the README.
