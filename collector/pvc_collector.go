package collector

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type podPvcLabelCollector struct {
	logger log.Logger
	desc   *prometheus.Desc
}

func (p podPvcLabelCollector) Update(ch chan<- prometheus.Metric) error {
	// Connect to Kubernetes API
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Errorf("cannot to create config: %v", err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Errorf("cannot to create clientset: %v", err)
	}

	// Get the namespace from the environment variable
	namespace := os.Getenv("NAMESPACE")

	persitentVolumeClaims := clientset.CoreV1().PersistentVolumeClaims(namespace)
	// List all pvcs in the namespace
	pvcList, err := persitentVolumeClaims.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Errorf("cannot to list pvcs: %v", err)
	}
	// Iterate over the pvc list and get owner reference and print pod labels for each pvc
	for _, pvc := range pvcList.Items {
		// Iterate over the owner reference list and get the pod name
		for _, ownerReference := range pvc.OwnerReferences {
			pod, err := clientset.CoreV1().Pods(pvc.Namespace).Get(context.TODO(), ownerReference.Name, v1.GetOptions{})
			if err != nil {
				fmt.Errorf("cannot to get pod: %v", err)
			}
			// Add labels from metadata to the metric
			ch <- prometheus.MustNewConstMetric(
				p.desc,
				prometheus.GaugeValue,
				1,
				pod.Name,
				pvc.Name,
				pod.Labels["jenkins/branch_name"],
				pod.Labels["jenkins/stage"],
				pod.Labels["jenkins/build"],
				pod.Labels["jenkins/job"],
				pod.Labels["jenkins/project"],
			)
		}

	}
	return nil
}

func NewPodPvcLabelCollector(logger log.Logger) (Collector, error) {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "pod", "pvc"),
		"pod_pvc_exporter: Export pod pvc labels from pod metadata",
		[]string{"pod", "pvc", "branch_name", "stage", "build", "job", "project"},
		nil,
	)
	return &podPvcLabelCollector{
		logger: logger,
		desc:   desc,
	}, nil
}

func init() {
	registerCollector("pod_pvc", defaultEnabled, NewPodPvcLabelCollector)
}
