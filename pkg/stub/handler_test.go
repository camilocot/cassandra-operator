package stub

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v1alpha1 "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/camilocot/cassandra-operator/pkg/util/probe"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
}

type MockCassandaCluster struct {
	mock.Mock
}

func (m *MockCassandaCluster) ReconcileService() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCassandaCluster) ReconcileMembers() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCassandaCluster) ReconcileStatefulset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCassandaCluster) ReconcileStatus() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCassandaCluster) SetDefaults() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCassandaCluster) FailedReconciliation(failedObjectName string, err error) error {
	_ = m.Called(failedObjectName, err)
	return fmt.Errorf("%s %v", failedObjectName, err)
}

func (suite *HandlerTestSuite) SetupTest() {
	// Run before each test...
}

func (suite *HandlerTestSuite) TestHandleWithNilObject() {
	ctx := context.TODO()
	event := sdk.Event{}
	assert.Nil(suite.T(), event.Object)

	handler := NewHandler()
	err := handler.Handle(ctx, event)
	assert.Nil(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandleWithDefault() {
	ctx := context.TODO()
	event := sdk.Event{Object: &v1alpha1.Cassandra{}}

	handler := NewHandler()
	err := handler.Handle(ctx, event)

	assert.Error(suite.T(), err)
}

func (suite *HandlerTestSuite) TestReconcileWithNilInput() {
	handler := NewHandler()
	err := handler.Reconcile(nil)
	assert.Error(suite.T(), err)
}

func (suite *HandlerTestSuite) TestReconcileWithValidInput() {
	cluster := new(MockCassandaCluster)

	cluster.On("SetDefaults").Return(false)
	cluster.On("ReconcileService").Return(nil)
	cluster.On("ReconcileMembers").Return(nil)
	cluster.On("ReconcileStatefulset").Return(nil)
	cluster.On("ReconcileStatus").Return(nil)

	handler := NewHandler()
	err := handler.Reconcile(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), probe.GetReady())
}

func (suite *HandlerTestSuite) TestReconcileWithServiceFailure() {
	err := errors.New("failed")
	cluster := new(MockCassandaCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("ReconcileService").Return(err)
	cluster.On("FailedReconciliation", "service", err).Return(nil)

	handler := NewHandler()
	err = handler.Reconcile(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "service failed", err.Error())
}

func (suite *HandlerTestSuite) TestReconcileWithMembersFailure() {
	err := errors.New("failed")
	cluster := new(MockCassandaCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("ReconcileService").Return(nil)
	cluster.On("ReconcileMembers").Return(errors.New("failed"))
	cluster.On("FailedReconciliation", "members", err).Return(nil)

	handler := NewHandler()
	err = handler.Reconcile(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "members failed", err.Error())
}

func (suite *HandlerTestSuite) TestReconcileWithStatefulsetFailure() {
	err := errors.New("failed")
	cluster := new(MockCassandaCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("ReconcileService").Return(nil)
	cluster.On("ReconcileMembers").Return(nil)
	cluster.On("ReconcileStatefulset").Return(errors.New("failed"))
	cluster.On("FailedReconciliation", "statefulset", err).Return(nil)

	handler := NewHandler()
	err = handler.Reconcile(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "statefulset failed", err.Error())
}

func (suite *HandlerTestSuite) TestReconcileWithStatusFailure() {
	err := errors.New("failed")
	cluster := new(MockCassandaCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("ReconcileService").Return(nil)
	cluster.On("ReconcileMembers").Return(nil)
	cluster.On("ReconcileStatefulset").Return(nil)
	cluster.On("ReconcileStatus").Return(errors.New("failed"))
	cluster.On("FailedReconciliation", "status", err).Return(nil)

	handler := NewHandler()
	err = handler.Reconcile(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "status failed", err.Error())
}

// Run test suite...
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
