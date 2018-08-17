package gimbalbench

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeleteTestNamespaces deletes all namespaces created by gimbalbench in a given cluster
func DeleteTestNamespaces(client *kubernetes.Clientset) error {
	nsList, err := ListTestNamespaces(client)
	if err != nil {
		return err
	}
	for _, ns := range nsList {
		if err := client.Core().Namespaces().Delete(ns.Name, &meta_v1.DeleteOptions{}); err != nil {
			return fmt.Errorf("error deleting namespace %q: %v", ns.Name, err)
		}
	}
	return nil
}

// ListTestNamespaces returns a list of namespaces that were created by gimbalbench in a given cluster.
func ListTestNamespaces(client *kubernetes.Clientset) ([]v1.Namespace, error) {
	var l []v1.Namespace
	nsList, err := client.Core().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, ns := range nsList.Items {
		if strings.HasPrefix(ns.Name, "gimbalbench") {
			l = append(l, ns)
		}
	}
	return l, nil
}
