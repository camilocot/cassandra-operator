package cassandra

import (
	"fmt"
	"reflect"

	apiv1alpha "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/camilocot/cassandra-operator/pkg/exec"
	"github.com/camilocot/cassandra-operator/pkg/util/probe"
	"github.com/sirupsen/logrus"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// Reconcile reconciles the cassandra cluster's state to the spec specified by cs
// by deploying the cassandra cluster,
func Reconcile(api *apiv1alpha.Cassandra) (err error) {

	probe.SetReady()

	api.SetDefaults()

	// Reconcile Service object
	err = reconcileService(api)
	if err != nil {
		return failedReconciliation("service", api, err)
	}

	// Reconcile Members
	err = reconcileMembers(api)
	if err != nil {
		return failedReconciliation("members", api, err)
	}

	// Reconcile StatefulSet object
	err = reconcileStatefulset(api)
	if err != nil {
		return failedReconciliation("statefulset", api, err)
	}

	// Reconcile Status object
	err = reconcileStatus(api)
	if err != nil {
		return failedReconciliation("status", api, err)
	}

	return nil

}

func reconcileService(api *apiv1alpha.Cassandra) (err error) {
	// Create the headless service if it doesn't exist
	existingSvc := Service(api)
	desiredSvc := Service(api)

	err = sdk.Get(existingSvc)
	if err != nil {
		err = sdk.Create(desiredSvc)
	} else {
		if !reflect.DeepEqual(existingSvc.Spec.Ports, desiredSvc.Spec.Ports) {
			existingSvc.Spec.Ports = desiredSvc.Spec.Ports
			err = sdk.Update(existingSvc)
		}
	}

	return err
}

func reconcileStatefulset(api *apiv1alpha.Cassandra) (err error) {

	existingSs := StatefulSet(api)
	desiredSs := StatefulSet(api)

	err = sdk.Get(existingSs)
	if err != nil {
		err = sdk.Create(desiredSs)
	} else {
		if !reflect.DeepEqual(existingSs.Spec, desiredSs.Spec) {
			existingSs.Spec = desiredSs.Spec
			err = sdk.Update(existingSs)
		}
	}
	return err
}

func reconcileStatus(api *apiv1alpha.Cassandra) (err error) {
	podNames, err := nodesForCassandra(api)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(podNames, api.Status.Members.Nodes) {
		api.Status.Members.Nodes = podNames
		err = sdk.Update(api)
	}

	api.Status.SetReadyCondition()
	return err
}

func reconcileMembers(api *apiv1alpha.Cassandra) (err error) {
	existing := StatefulSet(api)

	err = sdk.Get(existing)

	if err != nil {
		return err
	}

	desiredSize := api.Spec.Size
	existingSize := *existing.Spec.Replicas

	logrus.Infof("Existing size %v desiredSize: %v", existingSize, desiredSize)

	if existingSize <= desiredSize {
		return nil
	}

	podName := "cassandra-cluster-" + fmt.Sprint(existingSize-1)

	if existingSize != desiredSize+1 {
		return fmt.Errorf("statefulset could not be updated, instance decommission can only be done 1 by 1. Current replica: %v Desired size: %v", existingSize, desiredSize)
	}

	if api.Status.IsScaling() {
		return fmt.Errorf("statefulset is scaling")
	}

	api.Status.SetScalingDownCondition(int(desiredSize), int(existingSize))

	logrus.Infof("Start the decommission of %v", podName)

	out, err := exec.Command(api, podName, "nodetool", "decommission")

	logrus.Infof(out)
	logrus.Infof("Finished the decommission of %v", podName)
	return err
}

func failedReconciliation(object string, api *apiv1alpha.Cassandra, err error) error {

	api.Status.SetReason(err.Error())
	api.Status.SetPhase(apiv1alpha.ClusterPhaseFailed)

	return fmt.Errorf("[%s] API: %s Failed to reconcile %v %v", api.Namespace, api.Name, object, err)
}
