# Cluster bootstrap

Scaffold only. This directory will contain the ArgoCD ApplicationSet bootstrap
for the cluster.

Planned resources:

- ArgoCD installation/bootstrap configuration
- An `ApplicationSet` that discovers the GitOps applications
- An app-of-apps entry point for the `kargo-pattern` chart and Kargo resources
- Cluster-specific values and target revisions

No ArgoCD or Kargo manifests are installed from this repository yet.
