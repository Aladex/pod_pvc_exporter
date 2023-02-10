# Pod PVC Label Exporter

This is a Go-based package for a Prometheus exporter that exports the pod and PVC labels from a Kubernetes cluster. The exporter listens on port `8080` and provides metrics at the endpoint `/metrics`.

## Usage

The exporter can be run inside a Kubernetes cluster as a pod. The pod needs to have the `ClusterRole` `pod-pvc-label-exporter` and the `ClusterRoleBinding` `pod-pvc-label-exporter` to be able to list pods and PVCs in the cluster.

You need to provide the following environment variables to the pod:
- `NAMESPACE`: The namespace of the pod and PVCs.

## Metrics

The exporter provides the following metrics:
- `pod_pvc_labels`: A metric that provides the pod and PVC labels. The metric has the following labels:
  - `pod`: The name of the pod.
  - `pvc`: The name of the PVC,
  - `branch_name`: The branch name of the jenkins job that created the PVC.
  - `build`: The build number of the jenkins job that created the PVC.
  - `job_name`: The name of the jenkins job that created the PVC.
  - `stage`: The stage of the jenkins job that created the PVC.
  - `project`: The project of the jenkins job that created the PVC.

## Scaping

The exporter can be scraped by Prometheus using the following configuration:
```yaml
- job_name: 'pod-pvc-label-exporter'
  scrape_interval: 30s
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - <NAMESPACE>
  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_label_app]
    action: keep
    regex: pod-pvc-label-exporter
  - source_labels: [__meta_kubernetes_pod_container_port_number]
    action: keep
    regex: 8080
```

Don't forget to replace `<NAMESPACE>` with the namespace of the exporter pod.
