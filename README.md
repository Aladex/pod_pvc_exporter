# Pod PVC Label Exporter

This is a Go-based package for a Prometheus exporter that exports the pod and PVC labels from a Kubernetes cluster. The exporter listens on port `8080` and provides metrics at the endpoint `/metrics`.

## Usage

The exporter can be run inside a Kubernetes cluster as a pod. The pod needs to have the `ClusterRole` `pod-pvc-label-exporter` and the `ClusterRoleBinding` `pod-pvc-label-exporter` to be able to list pods and PVCs in the cluster.

You need to provide the following environment variables to the pod:
- `NAMESPACE`: The namespace of the pod and PVCs.


## Usage with TopoLVM

You can use the following queries to get information on the usage of a specific pod and PVC, including the amount of content used and available:

```
pvc_pod_info{exported_job!=""} * on(persistentvolumeclaim) group_left(pvc) kubelet_volume_stats_used_bytes{}
pvc_pod_info{exported_job!=""} * on(persistentvolumeclaim) group_left(pvc) kubelet_volume_stats_capacity_bytes{}
pvc_pod_info{exported_job!=""} * on(persistentvolumeclaim) group_left(pvc) kubelet_volume_stats_available_bytes{}

100 * 
  (pvc_pod_info{exported_job!=""} * on(persistentvolumeclaim) group_left(pvc) kubelet_volume_stats_used_bytes{})
    /
  (pvc_pod_info{exported_job!=""} * on(persistentvolumeclaim) group_left(pvc) kubelet_volume_stats_capacity_bytes{})
```

Here is an example of the output of these queries in Grafana:

| Date                | Job                          | Pod                        | Project                                | Used   | Capacity | Available | Usage    |
|---------------------|------------------------------|----------------------------|----------------------------------------|--------|----------|-----------|----------|
| 2023-03-10 14:35:15 | Documentation_build_job      | null                       | BUILD_api_documentation_doxygen      | 894    | 2.29 GiB | 13.7 GiB  | 14.3%    |
| 2023-03-09 12:47:32 | Backend_CI                   | backend-build-pod-1234-abc | Project_XYZ                            | 654    | 4.56 GiB | 21.4 GiB  | 3.0%     |
| 2023-03-08 10:15:44 | Frontend_Build               | frontend-build-pod-5678-df | Project_ABC                            | 158    | 1.34 GiB | 6.7 GiB   | 2.4%     |
| 2023-03-07 09:22:56 | Data_Processing_Job           | data-processing-pod-9012-gh | Data_Processing_Project                | 982    | 8.91 GiB | 21.5 GiB  | 11.0%    |


## Pod Annotations

The exporter can be configured using the following annotations on the pod:

- `jenkins/branch_name`: The branch name of the jenkins job that created the PVC
- `jenkins/stage`: The stage of the jenkins job that created the PVC
- `jenkins/build`: The build number of the jenkins job that created the PVC
- `jenkins/job`: The name of the jenkins job that created the PVC
- `jenkins/project`: The project of the jenkins job that created the PVC
- `jenkins/build_url`: The URL of the jenkins job that created the PVC

Here is an example of a pod configuration with these annotations:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-pvc-label-exporter
  namespace: jenkins
  annotations:
    jenkins/branch_name: BUILD_x86_64_linux_gcc
    jenkins/stage: Code_Check
    jenkins/build: "1234"
    jenkins/job: Build_Job
    jenkins/project: Project_A
    jenkins/build_url: https://jenkins.example.com/job/Build_Job/1234/
spec:
    containers:
    - name: ubuntu
      image: ubuntu
      command: ["/bin/bash", "-c", "sleep 3600"]
      volumeMounts:
        - name: workspace
          mountPath: /workspace
    volumes:
      - ephemeral:
          volumeClaimTemplate:
            spec:
              accessModes:
                - "ReadWriteOnce"
              resources:
                requests:
                  storage: "16Gi"
              storageClassName: "topolvm-provisioner"
        name: "workspace-lvm"
```
    

## Metrics

The exporter provides the following metrics:
- `pvc_pod_info`: A metric that provides the pod and PVC labels. The metric has the following labels:
  - `pod`: The name of the pod.
  - `persistentvolumeclaim`: The name of the PVC,
  - `branch_name`: The branch name of the jenkins job that created the PVC.
  - `build`: The build number of the jenkins job that created the PVC.
  - `job_name`: The name of the jenkins job that created the PVC.
  - `stage`: The stage of the jenkins job that created the PVC.
  - `project`: The project of the jenkins job that created the persistentvolumeclaim.

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

# Example Output

The following is an example output of the exporter:

```text
pvc_pod_info{branch_name="BUILD_x86_64_linux_gcc",build="1234",job="Build_Job",pod="build-pod-1234-abcde-fghij-klmno",project="Project_A",persistentvolumeclaim="pvc-workspace-1234-abcde-fghij-klmno",stage="Code_Check"} 1
pvc_pod_info{branch_name="BUILD_arm_android_clang",build="5678",job="Build_Job",pod="build-pod-5678-pqrst-uvwxy-zabcd",project="Project_B",persistentvolumeclaim="pvc-workspace-5678-pqrst-uvwxy-zabcd",stage="Unit_Test"} 1
pvc_pod_info{branch_name="PR_789",build="9",job="CI_Job_PR_android",pod="ci-pod-pr-android-789-9-12345-67890-abcd",project="Project_C",persistentvolumeclaim="pvc-workspace-pr-android-789-9-12345-67890-abcd",stage="Integration_Test"} 1
```