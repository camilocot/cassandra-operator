package cassandra

import (
	"fmt"
	"reflect"

	v1alpha1 "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/camilocot/cassandra-operator/pkg/exec"
	"github.com/sirupsen/logrus"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// Controller manages reconciliation of the Cassandra cluster
type Controller interface {
	ReconcileService() error
	ReconcileStatus() error
	ReconcileMembers() error
	ReconcileStatefulset() error
	SetDefaults() bool
	FailedReconciliation(string, error) error
}

// Cluster represents a Cassandra Cluster
type Cluster struct {
	Resource *v1alpha1.Cassandra

	Controller
}

// NewCassandraCluster creates a new Cassandra Cluster object
func NewCassandraCluster(c *v1alpha1.Cassandra) *Cluster {
	return &Cluster{Resource: c}
}

// ReconcileService reconciles the headless service
func (c Cluster) ReconcileService() (err error) {
	// Create the headless service if it doesn't exist
	r := c.Resource
	existingSvc := Service(r)
	desiredSvc := Service(r)

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

// ReconcileStatefulset reconciles the statefulset
func (c Cluster) ReconcileStatefulset() (err error) {

	r := c.Resource
	existingSs := StatefulSet(r)
	desiredSs := StatefulSet(r)

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

// ReconcileStatus reconciles the cluster status
func (c Cluster) ReconcileStatus() (err error) {
	r := c.Resource
	podNames, err := nodesForCassandra(r)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(podNames, r.Status.Members.Nodes) {
		r.Status.Members.Nodes = podNames
		err = sdk.Update(r)
	}

	r.Status.SetReadyCondition()
	return err
}

// ReconcileMembers reconciles the cluster member status
func (c Cluster) ReconcileMembers() (err error) {
	r := c.Resource
	existing := StatefulSet(r)

	err = sdk.Get(existing)

	// @TODO: Cluster not initialized, use Status to verify
	if err != nil {
		return nil
	}

	desiredSize := r.Spec.Size
	existingSize := existing.Status.Replicas

	logrus.Infof("Existing size %v desiredSize: %v", existingSize, desiredSize)

	if existingSize <= desiredSize {
		return nil
	}

	podName := "cassandra-cluster-" + fmt.Sprint(existingSize-1)

	if existingSize != desiredSize+1 {
		return fmt.Errorf("statefulset could not be updated, instance decommission can only be done 1 by 1. Current replica: %v Desired size: %v", existingSize, desiredSize)
	}

	if r.Status.IsScaling() {
		return fmt.Errorf("statefulset is scaling")
	}

	r.Status.SetScalingDownCondition(int(desiredSize), int(existingSize))

	logrus.Infof("Start the decommission of %v", podName)

	out, err := exec.Command(r, podName, "nodetool", "decommission")

	logrus.Infof(out)
	// @TODO: mark node as decommissioned
	logrus.Infof("Finished the decommission of %v", podName)
	return err
}

// FailedReconciliation set the cluster status to failed
func (c Cluster) FailedReconciliation(failedObjectName string, err error) error {

	c.Resource.Status.SetReason(err.Error())
	c.Resource.Status.SetPhase(v1alpha1.ClusterPhaseFailed)

	return fmt.Errorf("[%s] API: %s Failed to reconcile %v %v", c.Resource.Namespace, c.Resource.Name, failedObjectName, err)
}

// SetDefaults sets defaults resource values
func (c Cluster) SetDefaults() bool {
	return c.Resource.SetDefaults()
}
