package framework

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
)

func (f *Framework) ListNodes(cluster int) *v1.NodeList {
	nodeList, err := f.ClusterClients[cluster].CoreV1().Nodes().List(metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	return nodeList
}
