# Kargo

Scaffold only. This directory will contain the Kargo promotion pattern for the
`kargo-pattern` service.

Planned resources:

- `Project` for the service and its stages
- `Warehouse` watching the `ghcr.io/<owner>/kargo-patterns` image repository
- `Stage` definitions (`dev`, `staging`, `prod`)
- `PromotionTask` resources for updating the GitOps desired state
- `Freight` promotion rules and verification steps

The image published by the `image` workflow is the promotion artifact.

## Contract with ArgoCD

The `ApplicationSet` in `gitops/cluster/apps/` already expects these branches.
Promotions must:

- Push rendered manifests (plain YAML from the `chart/` chart) to the root of
  `deploy/app/<stage>`, creating the branch on the first promotion.
- Set `image.tag` explicitly for each stage, using `vX.Y.Z` or `sha-*`. Never
  promote `latest`, and do not rely on `Chart.AppVersion`.
- Set a unique ingress host for each stage. The chart default
  `kargo-pattern.localhost` is shared by all three namespaces, and the
  ingress-nginx admission webhook rejects the duplicate host and path.

Also required on the ArgoCD side:

- Add the `kargo.akuity.io/authorized-stage` annotation to the `ApplicationSet`
  template, not to the generated Applications, or the controller overwrites it.
- Check the first sync of each stage. `CreateNamespace=true` and the `Namespace`
  entry in the `AppProject` whitelist could not be exercised before the branches
  existed.
