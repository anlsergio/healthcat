package k8s

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ServiceRegistry interface {
	Add(name, url string)
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

	config, err := rest.InClusterConfig()
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
		case watch.Error:
			e.slogger.Errorf("Error listening to service events: %v", event.Object)
			break
		case watch.Added:
			e.addService(event.Object.(*v1.Service))
		case watch.Deleted:
			e.deleteService(event.Object.(*v1.Service))
		case watch.Modified:
			svc := event.Object.(*v1.Service)
			// TODO: This may not work as the commands may not come in order.
			// Consider using a single command channel in checker or
			// introduce a separate command for updating service
			e.deleteService(svc)
			e.addService(svc)
		default:
			e.slogger.Info("Ignoring unsupported event: %s", event.Type)
		}
	}
}

func (e *EventSource) addService(svc *v1.Service) {
	if !matchFilters(svc.Namespace, e.Namespaces, e.ExcludedNamespaces) {
		e.slogger.Debugf("Ignoring service %q in excluded namespace %q", svc.Name, svc.Namespace)
		return
	}
	schema := svc.ObjectMeta.Annotations["chc/schema"]
	if schema == "" {
		schema = "http"
	}

	path := svc.ObjectMeta.Annotations["chc/path"]
	if path == "" {
		path = "/healthz"
	}

	port := svc.Spec.Ports[0].Port
	e.slogger.Infof("Added service: %#v", svc.Name)
	e.Registry.Add(svc.Name, fmt.Sprintf("%s://%s:%d%s", schema, svc.Name, port, path))
}

func (e *EventSource) deleteService(svc *v1.Service) {
	e.Registry.Delete(svc.Name)
}

func matchFilters(namespace string, included, excluded []string) bool {
	for _, n := range excluded {
		if n == namespace {
			return false
		}
	}

	for _, n := range included {
		if n == namespace {
			return true
		}
	}

	return len(included) == 0
}
