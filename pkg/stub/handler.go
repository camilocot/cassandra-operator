package stub

import (
	"context"

	"github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"

	"github.com/camilocot/cassandra-operator/pkg/cassandra"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) (err error) {
	switch o := event.Object.(type) {
	case *v1alpha1.Cassandra:
		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		// All secondary resources must have the CR set as their OwnerReference for this to be the case
		if event.Deleted {
			return nil
		}
		err = cassandra.Reconcile(o)
		if err != nil {
			logrus.Errorf("Reconciliation error: %v", err)
		}
	}
	return err
}
