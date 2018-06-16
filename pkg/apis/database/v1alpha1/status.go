package v1alpha1

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
)

// ClusterPhase represents the status of the cluster
type ClusterPhase string

// ClusterConditionType represents the conditions by which the cluster has been
type ClusterConditionType string

const (
	// ClusterPhaseCreating represents creating cluster phase
	ClusterPhaseCreating = "Creating"
	// ClusterPhaseRunning represents running cluster phase
	ClusterPhaseRunning = "Running"
	// ClusterPhaseFailed represents failed cluster phase
	ClusterPhaseFailed = "Failed"

	// ClusterConditionAvailable represents available cluster condition
	ClusterConditionAvailable ClusterConditionType = "Available"
	// ClusterConditionScaling represents scaling cluster condition
	ClusterConditionScaling = "Scaling"
)

// ClusterStatus represents the current status of the cluster
type ClusterStatus struct {
	// Phase is the cluster running phase
	Phase  ClusterPhase `json:"phase"`
	Reason string       `json:"reason,omitempty"`

	// Condition keeps track of all cluster conditions, if they exist.
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// Size is the current size of the cluster
	Size int `json:"size"`
	// Members are the etcd members in the cluster
	Members MembersStatus `json:"members"`
	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion"`
	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion"`
}

// ClusterCondition represents one current condition of an cassandra cluster.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
// If a cluster is upgrading and encountered a problem that prevents the upgrade,
// the Upgrading condition's status will would be False and communicate the problem back.
type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// MembersStatus represents the status of the members of the cluster
type MembersStatus struct {
	// The nodes names are the same as the cassandra pod names
	Nodes []string `json:"nodes,omitempty"`
}

// Size is the number of the members of the cluster
func (ms *MembersStatus) Size() int {
	return len(ms.Nodes)
}

// IsFailed returns if the cluster is in a failed phase
func (cs *ClusterStatus) IsFailed() bool {
	if cs == nil {
		return false
	}
	return cs.Phase == ClusterPhaseFailed
}

// IsScaling returns if the cluster is in a scaling condition
func (cs *ClusterStatus) IsScaling() bool {
	if cs == nil {
		return false
	}
	c := cs.Conditions
	return c[len(c)-1].Type == ClusterConditionScaling
}

// SetPhase set the current phase of the cluster
func (cs *ClusterStatus) SetPhase(p ClusterPhase) {
	cs.Phase = p
}

// SetScalingDownCondition set scaling down condition
func (cs *ClusterStatus) SetScalingDownCondition(from, to int) {
	c := newClusterCondition(ClusterConditionScaling, v1.ConditionTrue, "Scaling down", scalingMsg(from, to))
	cs.setClusterCondition(*c)
}

// SetReadyCondition set ready condition
func (cs *ClusterStatus) SetReadyCondition() {
	c := newClusterCondition(ClusterConditionAvailable, v1.ConditionTrue, "Cluster available", "")
	cs.setClusterCondition(*c)
}
func (cs *ClusterStatus) setClusterCondition(c ClusterCondition) {
	pos, cp := getClusterCondition(cs, c.Type)
	if cp != nil &&
		cp.Status == c.Status && cp.Reason == c.Reason && cp.Message == c.Message {
		return
	}

	if cp != nil {
		cs.Conditions[pos] = c
	} else {
		cs.Conditions = append(cs.Conditions, c)
	}
}

func getClusterCondition(status *ClusterStatus, t ClusterConditionType) (int, *ClusterCondition) {
	for i, c := range status.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func newClusterCondition(condType ClusterConditionType, status v1.ConditionStatus, reason, message string) *ClusterCondition {
	now := time.Now().Format(time.RFC3339)
	return &ClusterCondition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	}
}

func scalingMsg(from, to int) string {
	return fmt.Sprintf("Current cluster size: %d, desired cluster size: %d", from, to)
}

// SetReason set the reason of the current status
func (cs *ClusterStatus) SetReason(r string) {
	cs.Reason = r
}
