package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultRepository = "gcr.io/google-samples/cassandra"
	// DefaultCassandraVersion default Cassandra docker image version
	DefaultCassandraVersion = "v13"

	// DefaultPartition default value for .spec.updateStrategy.rollingUpdate.partition
	DefaultPartition = 0
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraList is a list of Cassandra clusters
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra represents a Cassandra cluster
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CassandraSpec `json:"spec"`
	Status            ClusterStatus `json:"status,omitempty"`
}

// CassandraSpec contains the specification of the Cassandra cluster
type CassandraSpec struct {
	// Size is the expected size of the cassandra cluster.
	// The cassandra-operator will eventually make the size of the running
	// cluster equal to the expected size.
	Size int32 `json:"size"`
	// Repository is the name of the repository that hosts
	// cassandra container images.
	// That means, it should have exact same tags and the same meaning for the tags.
	//
	// By default, it is `gcr.io/google-samples/cassandra`.
	Repository string `json:"repository,omitempty"`

	// Version is the expected version of the cassandra image
	// The cassandra-operator will eventually make the cassandra cluster version
	// equal to the expected version.
	//
	// If version is not set, default is "v13".
	Version string `json:"version,omitempty"`
	// Partition is the expected number of pods that will be kept with the
	// current version
	//
	// If partition is not set, default is 0.
	Partition int32 `json:"partition,omitempty"`

	StorageClassName string `json:"storageClassName"`

	// List of environment variables to set in the cassandra container.
	// This is used to configure cassandra process. Cassandra cluster cannot be created, when
	// bad environement variables are provided.
	// This field cannot be updated.
	CassandraEnv []v1.EnvVar `json:"cassandraEnv,omitempty"`
}

func (c *Cassandra) addEnvVar(name string, value string) {
	cs := &c.Spec

	for _, v := range cs.CassandraEnv {
		if v.Name == name {
			return
		}
	}

	cs.CassandraEnv = append(cs.CassandraEnv, v1.EnvVar{
		Name:  name,
		Value: value,
	})

}

// SetDefaults set the defaults values for the Cassandra cluster. Set docker image and
// the following env variables: CASSANDRA_SEEDS, MAX_NEWSIZE and MAX_HEAP_SIZE
func (c *Cassandra) SetDefaults() bool {
	changed := false
	cs := &c.Spec

	if len(cs.Repository) == 0 {
		cs.Repository = defaultRepository
		changed = true
	}

	if len(cs.Version) == 0 {
		cs.Version = DefaultCassandraVersion
		changed = true
	}

	c.addEnvVar("CASSANDRA_SEEDS", c.Name+"-0."+c.Name+"-unready."+c.Namespace+".svc.cluster.local")
	c.addEnvVar("MAX_HEAP_SIZE", "512M")
	c.addEnvVar("MAX_NEWSIZE", "100M")

	return changed
}
