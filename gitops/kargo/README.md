# Kargo

Scaffold only. This directory will contain the Kargo promotion pattern for the
`kargo-pattern` service.

Planned resources:

- `Project` for the service and its stages
- `Warehouse` watching the `ghcr.io/<owner>/kargo-patterns` image repository
- `Stage` definitions (for example `dev`, `staging`, `prod`)
- `PromotionTask` resources for updating the GitOps desired state
- `Freight` promotion rules and verification steps

The image published by the `image` workflow is the promotion artifact.
