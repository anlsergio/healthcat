package k8s

import (
	"context"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Start starts the loop
func Start() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	watch, err := clientset.CoreV1().Pods("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		obj := event.Object
		if pod, ok := obj.(*v1.Pod); ok {
			log.Printf("event: %v, pod %v in %v", event.Type, pod.Name, pod.Status.Phase)
		} else {
			log.Printf("Expected pod; got %T", obj)
		}
	}
	return nil
}
