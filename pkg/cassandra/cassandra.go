package cassandra

import (
	"github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/kubernetes/pkg/util/pointer"
)

// StatefulSet returns a cassandra StatefulSet object
func StatefulSet(api *v1alpha1.Cassandra) *appsv1.StatefulSet {
	labels := labelsForCassandra(api.Name)
	replicas := api.Spec.Size
	partition := api.Spec.Partition
	storageClass := api.Spec.StorageClassName
	env := append(api.Spec.CassandraEnv, v1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	})

	stateful := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      api.Name,
			Namespace: api.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: api.Name + "-unready",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					Partition: &partition,
				},
			},
			RevisionHistoryLimit: pointer.Int32Ptr(10),
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: &storageClass,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cassandra",
					},
				},
			},

			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "cassandra",
							Image: api.Spec.Repository + ":" + api.Spec.Version,
							Env:   env,
							Ports: []v1.ContainerPort{
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
								},
							},
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Add: []v1.Capability{"IPC_LOCK"},
								},
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/bash", "-c", "/ready-probe.sh"},
									},
								},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
							},
							Lifecycle: &v1.Lifecycle{
								PreStop: &v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "-c", "nodetool", "drain"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	addOwnerRefToObject(stateful, asOwner(api))
	return stateful
}

// Service returns a cassandra Service object
func Service(api *v1alpha1.Cassandra) *v1.Service {

	labels := labelsForCassandra(api.Name)

	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      api.Name + "-unready",
			Labels:    labels,
			Namespace: api.Namespace,
			// it will return IPs even of the unready pods. Bootstraping a new cluster need it
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "cql",
					Port:       9042,
					TargetPort: intstr.FromInt(9042),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector:  labels,
			ClusterIP: "None",
			Type:      v1.ServiceTypeClusterIP,
		},
	}
	addOwnerRefToObject(svc, asOwner(api))
	return svc
}

// labelsForCassadnra returns the labels for selecting the resources
// belonging to the given casandra CR name.
func labelsForCassandra(name string) map[string]string {
	return map[string]string{
		"app":          "cassandra",
		"cassandra_cr": name,
	}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// asOwner returns an OwnerReference set as the cassandra CR
func asOwner(api *v1alpha1.Cassandra) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.APIVersion,
		Kind:       api.Kind,
		Name:       api.Name,
		UID:        api.UID,
		Controller: &trueVar,
	}
}

// podList returns a v1.PodList object
func podList() *v1.PodList {
	return &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []v1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func nodesForCassandra(api *v1alpha1.Cassandra) ([]string, error) {
	podList := podList()
	labelSelector := labels.SelectorFromSet(labelsForCassandra(api.Name)).String()
	listOps := &metav1.ListOptions{LabelSelector: labelSelector}
	err := sdk.List(api.Namespace, podList, sdk.WithListOptions(listOps))
	if err != nil {
		return nil, err
	}
	podNames := getPodNames(podList.Items)
	return podNames, nil
}
