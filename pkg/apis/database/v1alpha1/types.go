package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cassandra `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CassandraSpec   `json:"spec"`
	Status            CassandraStatus `json:"status,omitempty"`
}

type CassandraSpec struct {
	// Size is the size of the cassandra Statefulset
	Size int32 `json:"size"`
}
type CassandraStatus struct {
	// Nodes are the names of the nodes of the cassandra pods
	Nodes []string `json:"nodes"`
}
