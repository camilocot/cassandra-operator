package cassandra

import (
	"testing"

	api "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
)

func TestUpdateStatefulset(t *testing.T) {
	c := api.Cassandra{
		Spec: api.CassandraSpec{
			Partition:  1,
			Repository: "repo1",
			Size:       1,
		},
	}
	s := statefulsetForCassandra(&c)

	c.Spec.Partition = 2
	s = updateStatefulset(&c, s)
	if c.Spec.Partition != *s.Spec.UpdateStrategy.RollingUpdate.Partition {
		t.Errorf("Partition wasn't updated, expected %v but was %v", c.Spec.Partition, *s.Spec.UpdateStrategy.RollingUpdate.Partition)
	}
}
