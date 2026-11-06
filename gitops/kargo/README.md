# Kargo

Release-train promotion pipeline for the `kargo-pattern` service, in the Kargo
Project `release-train`. ArgoCD applies this directory through the `kargo-project`
Application.

This directory holds only the pipeline. The application configuration it promotes
lives in [`../configs/demo/app-values/`](../configs/demo/app-values/).

```
          app Warehouse: image + chart + demo/app-values/default.yaml
                                  |
                                  v
dev -> staging-gate -> staging -> prod-gate -> prod
 ^                       ^                      ^
values-dev         values-staging          values-prod
```

## Stages

| Stage | Kind | `app` Freight | `values-<stage>` Freight |
|---|---|---|---|
| `dev` | regular | automatic | automatic |
| `staging-gate` | control flow | automatic | none |
| `staging` | regular | **manual** | automatic |
| `prod-gate` | control flow | automatic | none |
| `prod` | regular | **manual** | automatic |

A gate has no promotion process. Kargo immediately marks as verified any Freight
verified in the previous Stage, so the gate collects every candidate for the next
manual promotion. No `Promotion` runs for a gate.

## Warehouses

- `app`: the released image (`vX.Y.Z` tags only), the released chart and
  `gitops/configs/demo/app-values/default.yaml`. They produce one Freight and travel together
  through every Stage.
- `values-dev`, `values-staging`, `values-prod`: each watches one file in
  `gitops/configs/demo/app-values/` and feeds one Stage only, so environment values are never
  promoted.

A Warehouse needs a release to produce `app` Freight: until the image and the chart
are published, it reports an error for the missing artifact.

## What a promotion does

`promotion-task.yaml` renders the released chart with these values, lowest
precedence first:

1. the defaults in the chart
2. `gitops/configs/demo/app-values/default.yaml`, at the commit pinned in the `app` Freight
3. `gitops/configs/demo/app-values/<stage>.yaml`, at the commit pinned in the `values-<stage>`
   Freight
4. `image.tag` from the `app` Freight

It then replaces the content of the `deploy/app/<stage>` branch and syncs the
Application:

```
manifests/     rendered chart, installed by ArgoCD
freight.yaml   artifacts currently deployed to the Stage
```

`freight.yaml` lists the promotion, the image tag and digest, the chart version and
the commits of both values files. ArgoCD does not read it.

## Promoting

Promote from the gate, which holds every candidate:

```sh
kargo login http://kargo.localhost:8080 --admin
kargo get freight --project release-train
kargo promote --project release-train --freight <name> --downstream-from staging-gate
kargo promote --project release-train --freight <name> --downstream-from prod-gate
```

## Things that behave differently than they look

**Auto-promotion is enabled on `staging` and `prod`, and the UI shows it.** Kargo
enables it per Stage, not per Freight request, and the values Freight needs it. The
`app` request stays manual because of its `MatchUpstream` policy: it selects only
Freight currently in the upstream Stage, and a gate never holds any. As a result,
`app` Freight is not offered as a candidate for these Stages, and a manual promotion
records an auto-promotion hold on the `app` request. The hold is harmless and does not
affect the values Freight.

**A commit to `gitops/configs/demo/app-values/prod.yaml` reaches prod automatically**, without
passing the `prod-gate`. Treat that file like a production change.

**The first Freight on a Stage does not deploy anything.** A Stage renders only when
both of its requests are filled, so the first promotion completes without pushing
and the second one renders. On `dev`, if the `app` Freight arrives first, it counts
as verified before the first deployment happens.

## Adding a service to the train

Add a Warehouse for the service, and a `requestedFreight` entry for it on every Stage.
The Promotion Task and the ArgoCD Application also need a per-service variant. Kargo
creates one `Promotion` per Freight, so a train with several services takes one
promotion per service: there is no atomic promotion of the whole gate.
