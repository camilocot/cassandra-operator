package cassandra

import (
	"fmt"
	"reflect"

	"github.com/Sirupsen/logrus"
	api "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/camilocot/cassandra-operator/pkg/exec"
	"github.com/camilocot/cassandra-operator/pkg/util/probe"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Reconcile reconciles the cassandra cluster's state to the spec specified by cs
// by deploying the cassandra cluster,
func Reconcile(cassandra *api.Cassandra) (err error) {
	probe.SetReady()

	cassandra = cassandra.DeepCopy()
	cassandra.SetDefaults()
	cassandra.Status.SetReadyCondition()
	// Create the headless service if it doesn't exist
	svc := headLessServiceUnreadyForCassandra(cassandra)

	err = sdk.Get(svc)
	if err != nil {
		err = sdk.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create headless unready service: %v", err)
		}
	}

	// Create the statefulset if it doesn't exist
	s := statefulsetForCassandra(cassandra)
	err = sdk.Create(s)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create statefulset: %v", err)
	}

	// Ensure the statefulset size is the same as the spec
	err = sdk.Get(s)
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %v", err)
	}

	size := cassandra.Spec.Size

	if *s.Spec.Replicas != size {
		if cassandra.Status.IsScaling() {
			return fmt.Errorf("A scaling operation is in progress, can't start another")
		}
		if *s.Spec.Replicas > size {
			err = removeOneMember(cassandra, *s.Spec.Replicas)
			if err != nil {
				return err
			}
		}
	}

	stateful := updateStatefulset(cassandra, s)

	err = sdk.Update(stateful)
	if err != nil {
		return fmt.Errorf("failed to update statefulset: %v", err)
	}

	podNames, err := nodesForCassandra(cassandra)
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	if !reflect.DeepEqual(podNames, cassandra.Status.Members.Nodes) {
		cassandra.Status.Members.Nodes = podNames
		err := sdk.Update(cassandra)
		if err != nil {
			return fmt.Errorf("failed to update cassandra status: %v", err)
		}
	}

	return err

}

func updateStatefulset(c *api.Cassandra, s *appsv1.StatefulSet) *appsv1.StatefulSet {
	stateful := s.DeepCopy()
	size := c.Spec.Size
	image := c.Spec.Repository + ":" + c.Spec.Version
	partition := c.Spec.Partition

	stateful.Spec.Replicas = &size
	stateful.Spec.Template.Spec.Containers[0].Image = image
	stateful.Spec.UpdateStrategy.RollingUpdate.Partition = &partition

	return stateful
}

func removeOneMember(c *api.Cassandra, currentReplicas int32) error {
	size := c.Spec.Size
	if currentReplicas != size+1 {
		return fmt.Errorf("statefulset could not be updated, instance decommission can only be done 1 by 1. Current replica: %v Size: %v", currentReplicas, size)
	}

	err := removeMember(c)
	if err != nil {
		c.Status.SetReason(err.Error())
		c.Status.SetPhase(api.ClusterPhaseFailed)
		return err
	}

	c.Status.SetReadyCondition()
	return nil
}

func removeMember(c *api.Cassandra) error {
	size := c.Spec.Size
	c.Status.SetScalingDownCondition(c.Status.Members.Size(), int(size))
	logrus.Infof("Start the decommission of cassandra-cluster-" + fmt.Sprint(size))
	out, _ := exec.ExecCommand(c, "cassandra-cluster-"+fmt.Sprint(size), "nodetool", "decommission")
	logrus.Infof("Finished the decommission of cassandra-cluster-" + fmt.Sprint(size))

	logrus.Infof(out)

	return nil
}
