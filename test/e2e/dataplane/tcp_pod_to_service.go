package dataplane

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/submariner-io/submariner/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("[dataplane] Basic Pod to Service tests across clusters without discovery", func() {
	f := framework.NewDefaultFramework("dataplane-p2s-nd")

	It(
		"Should be able to perform a Pod to Service TCP connection and exchange data between different clusters", func() {
			bNodes := f.ListNodes(framework.ClusterB)
			aNodes := f.ListNodes(framework.ClusterA)
			for _, nodeB := range bNodes.Items {
				for _, nodeA := range aNodes.Items {
					nodeA := nodeA
					nodeB := nodeB
					By(fmt.Sprintf("Testing to exchange data between nodes(%s<->%s)", nodeB.Name, nodeA.Name))
					testPod2ServiceTCP(f, nodeB.Name, nodeA.Name)
				}
			}
		})

})

func testPod2ServiceTCP(f *framework.Framework, nodeB string, nodeA string) {
	defer GinkgoRecover()
	listenerUUID := string(uuid.NewUUID())
	connectorUUID := string(uuid.NewUUID())

	By(fmt.Sprintf("Creating a listener pod in cluster B(%s), which will wait for a handshake over TCP", f.ClusterContext[framework.ClusterB]))
	listenerPod := f.CreateTCPCheckListenerPod(framework.ClusterB, nodeB, listenerUUID)
	framework.Logf("Listener Pod is on  Node: %v", listenerPod.Spec.NodeName)

	By("Pointing a service ClusterIP to the listener pod in cluster B")
	service := f.CreateTCPService(framework.ClusterB, listenerPod.Labels[framework.TestAppLabel], framework.TestPort)
	framework.Logf("Service for listener pod has ClusterIP: %v", service.Spec.ClusterIP)

	By(fmt.Sprintf("Creating a connector pod in cluster A(%s), which will attempt the specific UUID handshake over TCP", f.ClusterContext[framework.ClusterA]))
	connectorPod := f.CreateTCPCheckConnectorPod(framework.ClusterA, listenerPod, service.Spec.ClusterIP, nodeA, connectorUUID)
	connectorPod = f.WaitForPodToBeReady(connectorPod, framework.ClusterA)
	framework.Logf("Connector Pod is on  Node: %v", connectorPod.Spec.NodeName)

	By("Waiting for the listener pod to exit with code 0, returning what listener sent")
	exitStatusL, exitMessageL := f.WaitForPodFinishStatus(listenerPod, framework.ClusterB)
	framework.Logf("Listener output:\n%s", exitMessageL)
	Expect(exitStatusL).To(Equal(int32(0)))

	By("Waiting for the connector pod to exit with code 0, returning what connector sent")
	exitStatusC, exitMessageC := f.WaitForPodFinishStatus(connectorPod, framework.ClusterA)
	framework.Logf("Connector output\n%s", exitMessageC)
	Expect(exitStatusC).To(Equal(int32(0)))

	By("Verifying what the pods sent to each other contain the right UUIDs")
	Expect(exitMessageL).To(ContainSubstring(connectorUUID))
	Expect(exitMessageC).To(ContainSubstring(listenerUUID))
}
