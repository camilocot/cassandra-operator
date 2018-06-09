package cassandra

import (
	"fmt"
	"reflect"

	api "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/camilocot/cassandra-operator/pkg/util/probe"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Reconcile reconciles the cassandra cluster's state to the spec specified by cs
// by deploying the cassandra cluster,
func Reconcile(cassandra *api.Cassandra) (err error) {
	updated := false
	probe.SetReady()

	cassandra = cassandra.DeepCopy()
	cassandra.SetDefaults()
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
	stateful := statefulsetForCassandra(cassandra)
	err = sdk.Create(stateful)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create statefulset: %v", err)
	}

	// Ensure the statefulset size is the same as the spec
	err = sdk.Get(stateful)
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %v", err)
	}

	size := cassandra.Spec.Size

	if *stateful.Spec.Replicas != size {
		if *stateful.Spec.Replicas > size {
			if *stateful.Spec.Replicas != size+1 {
				return fmt.Errorf("statefulset could not be updated, instance decommission can only be done 1 by 1 ")
			}
			logrus.Infof("Start the decommission of cassandra-cluster-" + fmt.Sprint(size))
			// @TODO: sync commands
			out, _ := ExecCommand(cassandra, "cassandra-cluster-"+fmt.Sprint(size), "nodetool", "decommission")
			logrus.Infof("Finished the decommission of cassandra-cluster-" + fmt.Sprint(size))

			fmt.Println(out)
		}
		stateful.Spec.Replicas = &size
		updated = true
	}

	image := cassandra.Spec.Repository + ":" + cassandra.Spec.Version

	if stateful.Spec.Template.Spec.Containers[0].Image != image {
		stateful.Spec.Template.Spec.Containers[0].Image = image
		updated = true
	}

	partition := cassandra.Spec.Partition

	if *stateful.Spec.UpdateStrategy.RollingUpdate.Partition != partition {
		stateful.Spec.UpdateStrategy.RollingUpdate.Partition = &partition
		updated = true
	}

	if updated {
		err = sdk.Update(stateful)
		if err != nil {
			return fmt.Errorf("failed to update statefulset: %v", err)
		}
	}

	podNames, err := nodesForCassandra(cassandra)
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	if !reflect.DeepEqual(podNames, cassandra.Status.Nodes) {
		cassandra.Status.Nodes = podNames
		err := sdk.Update(cassandra)
		if err != nil {
			return fmt.Errorf("failed to update cassandra status: %v", err)
		}
	}

	return err

}
