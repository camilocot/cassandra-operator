package cassandra

import (
	"testing"

	v1alpha1 "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCassandra() *v1alpha1.Cassandra {
	return &v1alpha1.Cassandra{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cassandra",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: v1alpha1.CassandraSpec{
			Size:             2,
			Repository:       "repository",
			Version:          "version",
			Partition:        1,
			StorageClassName: "storageClassName",
			CassandraEnv: []v1.EnvVar{
				{
					Name:  "Env1",
					Value: "Value1",
				},
				{
					Name:  "Env2",
					Value: "Value2",
				},
			},
		},
	}
}

func TestStatefulSet(t *testing.T) {
	cs := NewCassandra()
	st := StatefulSet(cs)
	pod := st.Spec.Template.Spec
	trueVar := true

	assert.Equal(t, cs.Name, st.ObjectMeta.Name)
	assert.Equal(t, cs.Namespace, st.ObjectMeta.Namespace)

	assert.Equal(t, cs.Name+"-unready", st.Spec.ServiceName)
	assert.Equal(t, cs.Spec.Size, *st.Spec.Replicas)
	assert.Equal(t, appsv1.OrderedReadyPodManagement, st.Spec.PodManagementPolicy)

	assert.Equal(t, appsv1.StatefulSetUpdateStrategyType(appsv1.RollingUpdateStatefulSetStrategyType), st.Spec.UpdateStrategy.Type)
	assert.Equal(t, cs.Spec.Partition, *st.Spec.UpdateStrategy.RollingUpdate.Partition)

	assert.Equal(t, 1, len(pod.Containers))

	c := pod.Containers[0]
	assert.Equal(t, cs.Spec.Repository+":"+cs.Spec.Version, c.Image)
	assert.Equal(t, "cassandra", c.Name)

	assert.Equal(t, []v1.ContainerPort{
		{
			Name:          "cql",
			ContainerPort: 9042,
		},
		{
			Name:          "intra-node",
			ContainerPort: 7001,
		},
		{
			Name:          "jmx",
			ContainerPort: 7099,
		}}, c.Ports)

	assert.Equal(t, []v1.Capability{"IPC_LOCK"}, c.SecurityContext.Capabilities.Add)

	assert.Equal(t, []string{"/bin/bash", "-c", "/ready-probe.sh"}, c.ReadinessProbe.Handler.Exec.Command)
	assert.Equal(t, int32(15), c.ReadinessProbe.InitialDelaySeconds)
	assert.Equal(t, int32(5), c.ReadinessProbe.TimeoutSeconds)
	assert.Equal(t, []string{"/bin/sh", "-c", "nodetool", "drain"}, c.Lifecycle.PreStop.Exec.Command)
	assert.Equal(t, append(cs.Spec.CassandraEnv, v1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		}}),
		c.Env)

	assert.Equal(t, 1, len(st.Spec.VolumeClaimTemplates))
	vct := st.Spec.VolumeClaimTemplates[0]
	assert.Equal(t, []v1.PersistentVolumeAccessMode{"ReadWriteOnce"}, vct.Spec.AccessModes)
	assert.Equal(t, cs.Spec.StorageClassName, *vct.Spec.StorageClassName)
	assert.Equal(t, v1.ResourceList{
		v1.ResourceStorage: resource.MustParse("1Gi"),
	}, vct.Spec.Resources.Requests)

	assert.Equal(t, 1, len(st.OwnerReferences))
	assert.Equal(t, metav1.OwnerReference{
		APIVersion: cs.APIVersion,
		Kind:       cs.Kind,
		Name:       cs.Name,
		UID:        cs.UID,
		Controller: &trueVar,
	}, st.OwnerReferences[0])
}
