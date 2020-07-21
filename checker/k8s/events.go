package k8s

import (
	"context"
	"fmt"
	"os"
	"path"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ServiceRegistry interface {
	Add(name string)
	Delete(name string)
}

type EventSource struct {
	Logger             *zap.Logger
	Namespaces         []string
	ExcludedNamespaces []string
	Registry           ServiceRegistry

	clientset *kubernetes.Clientset
	slogger   *zap.SugaredLogger
}

// Start starts the loop
func (e *EventSource) Start() error {
	e.slogger = e.Logger.Sugar()

	config, err := getConfig()
	if err != nil {
		return err
	}

	e.clientset, err = kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	go e.Run()

	return nil
}

func (e *EventSource) Run() {
	serviceWatch, err := e.clientset.CoreV1().Services("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		e.slogger.Errorf("Error while watching services: %v", err)
		return
	}

	for event := range serviceWatch.ResultChan() {
		switch event.Type {
		case watch.Added:
			svc := event.Object.(*v1.Service)

			e.slogger.Infof("Added service: %#v", svc)
			e.Registry.Add(fmt.Sprintf("http://%s:%d", svc.Name, svc.Spec.Ports[0].Port))
		case watch.Deleted:
			svc := event.Object.(*v1.Service)
			e.Registry.Delete(fmt.Sprintf("http://%s:%d", svc.Name, svc.Spec.Ports[0].Port))
		}
	}
}

func getConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	if err != rest.ErrNotInCluster {
		return nil, err
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", path.Join(homedir, ".kube", "config"))
}
