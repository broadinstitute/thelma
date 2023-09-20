package sherlock

import (
	"github.com/pkg/errors"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type environmentsSuite struct {
	suite.Suite
}

func TestEnvironmentsSuite(t *testing.T) {
	suite.Run(t, new(environmentsSuite))
}

func (suite *environmentsSuite) TestDeleteNonExistingEnvironment() {
	mockStateReadWriter := mocks.NewStateReadWriter(suite.T())
	mockStateReadWriter.AssertNotCalled(suite.T(), "DeleteEnvironments", mock.AnythingOfType("[]terra.Environment"))
	mockState := state{
		sherlock: mockStateReadWriter,
	}

	testEnvironments := newEnvironmentsView(&mockState)

	err := testEnvironments.Delete("blah")
	suite.Assert().Error(err)
}

func (suite *environmentsSuite) TestDeleteExistingEnvironment() {
	mockEnvironment := &environment{
		destination: destination{
			name: "existing-env",
		},
	}
	mockEnvironments := make(map[string]*environment)
	mockEnvironments["existing-env"] = mockEnvironment
	mockStateReadWriter := mocks.NewStateReadWriter(suite.T())
	mockStateReadWriter.On("DeleteEnvironments", mock.AnythingOfType("[]terra.Environment")).Return([]string{"existing-env"}, nil)
	mockState := state{
		sherlock:     mockStateReadWriter,
		environments: mockEnvironments,
	}
	testEnvironments := newEnvironmentsView(&mockState)
	err := testEnvironments.Delete("existing-env")
	suite.Assert().NoError(err)
}

func (suite *environmentsSuite) TestDeleteEnvironmentError() {
	mockEnvironment := &environment{
		destination: destination{
			name: "existing-env",
		},
	}
	mockEnvironments := make(map[string]*environment)
	mockEnvironments["existing-env"] = mockEnvironment
	mockStateReadWriter := mocks.NewStateReadWriter(suite.T())
	mockStateReadWriter.On("DeleteEnvironments", mock.AnythingOfType("[]terra.Environment")).Return(nil, errors.Errorf("some error"))
	mockState := state{
		sherlock:     mockStateReadWriter,
		environments: mockEnvironments,
	}
	testEnvironments := newEnvironmentsView(&mockState)
	err := testEnvironments.Delete("existing-env")
	suite.Assert().ErrorContains(err, "some error")
}
