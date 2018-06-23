package stub

import (
	"context"
	"fmt"

	"github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/camilocot/cassandra-operator/pkg/util/probe"

	"github.com/camilocot/cassandra-operator/pkg/cassandra"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
)

// NewHandler is executed when in Cassandra resource events (create, delete and update)
func NewHandler() *CassandraHandler {
	return &CassandraHandler{}
}

// CassandraHandler type
type CassandraHandler struct {
	sdk.Handler
}

// Handle reacts to events and outputs actions.
func (h *CassandraHandler) Handle(ctx context.Context, event sdk.Event) (err error) {
	switch o := event.Object.(type) {
	case *v1alpha1.Cassandra:
		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		if event.Deleted {
			return nil
		}
		err = h.Reconcile(cassandra.NewCassandraCluster(o))
		if err != nil {
			logrus.Errorf("Reconciliation error: %v", err)
		}
	}
	return err
}

// Reconcile reconciles the cassandra cluster's state to the spec specified by cs
// by deploying the cassandra cluster,
func (h *CassandraHandler) Reconcile(c cassandra.Controller) (err error) {
	if c == nil {
		return fmt.Errorf("controller cannot be nil")
	}

	probe.SetReady()

	c.SetDefaults()

	// Reconcile Service object
	err = c.ReconcileService()
	if err != nil {
		return c.FailedReconciliation("service", err)
	}

	// Reconcile Members
	err = c.ReconcileMembers()
	if err != nil {
		return c.FailedReconciliation("members", err)
	}

	// Reconcile StatefulSet object
	err = c.ReconcileStatefulset()
	if err != nil {
		return c.FailedReconciliation("statefulset", err)
	}

	// Reconcile Status object
	err = c.ReconcileStatus()
	if err != nil {
		return c.FailedReconciliation("status", err)
	}

	return nil

}
