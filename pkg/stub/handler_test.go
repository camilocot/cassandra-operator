package stub

import (
	"context"
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
	args := m.Called(failedObjectName, err)
	return args.Error(0)
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

// Run test suite...
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
