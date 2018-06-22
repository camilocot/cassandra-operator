package stub

import (
	"context"
	"testing"

	v1alpha1 "github.com/camilocot/cassandra-operator/pkg/apis/database/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
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

// Run test suite...
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
