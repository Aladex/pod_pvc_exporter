package collector

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type podPvcLabelCollector struct {
	logger log.Logger
	desc   *prometheus.Desc
}

func (p podPvcLabelCollector) Update(ch chan<- prometheus.Metric) error {
	// Connect to Kubernetes API
	var config *rest.Config
	// Check if running inside a cluster else use config from home directory
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Errorf("cannot to create config: %v", err)
		}
	} else {
		home := os.Getenv("HOME")
		config, err = clientcmd.BuildConfigFromFlags("", home+"/.kube/config")
		if err != nil {
			fmt.Errorf("cannot to create config: %v", err)
		}
	}

	// Set namespace from environment variable, if not set use default 'jenkins'
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "jenkins"
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Errorf("cannot to create clientset: %v", err)
	}

	// Get the namespace from the environment variable

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
			// Add labels from pod annotations to the metric as labels
			ch <- prometheus.MustNewConstMetric(
				p.desc,
				prometheus.GaugeValue,
				1,
				pod.Name,
				pvc.Name,
				pod.Annotations["jenkins/branch_name"],
				pod.Annotations["jenkins/stage"],
				pod.Annotations["jenkins/build"],
				pod.Annotations["jenkins/job"],
				pod.Annotations["jenkins/project"],
				pod.Annotations["jenkins/build_url"],
			)
		}

	}
	return nil
}

func NewPodPvcLabelCollector(logger log.Logger) (Collector, error) {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName("pvc", "pod", "info"),
		"pod_pvc_exporter: Export pod pvc labels from pod metadata",
		[]string{"pod", "persistentvolumeclaim", "branch_name", "stage", "build", "job", "project", "build_url"},
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
