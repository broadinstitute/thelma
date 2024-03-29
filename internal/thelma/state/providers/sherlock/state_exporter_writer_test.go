package sherlock_test

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/pkg/errors"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/sherlock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type sherlockStateWriterSuite struct {
	suite.Suite
	state terra.State
}

func TestSherlockStateWriterProviderSuite(t *testing.T) {
	suite.Run(t, new(sherlockStateWriterSuite))
}

func (s *sherlockStateWriterSuite) SetupSuite() {
	s.state = constructFakeState(s.T())
}

func (suite *sherlockStateWriterSuite) TestSuccessfulStateWriter() {
	mockStateWriter := mocks.NewStateWriter(suite.T())
	mockStateWriter.On("WriteClusters", mock.AnythingOfType("[]terra.Cluster")).Return(nil)
	mockStateWriter.On("WriteEnvironments", mock.AnythingOfType("[]terra.Environment")).Return([]string{"environment-a"}, nil)

	stateWriter := sherlock.NewSherlockStateWriter(suite.state, mockStateWriter)
	err := stateWriter.WriteClusters()
	suite.Assert().NoError(err)
	mockStateWriter.AssertNumberOfCalls(suite.T(), "WriteClusters", 1)

	err = stateWriter.WriteEnvironments()
	suite.Assert().NoError(err)
	mockStateWriter.AssertNumberOfCalls(suite.T(), "WriteEnvironments", 2)
}

func (suite *sherlockStateWriterSuite) TestErrorWriteEnvironments() {
	mockStateWriter := mocks.NewStateWriter(suite.T())
	mockStateWriter.On("WriteEnvironments", mock.AnythingOfType("[]terra.Environment")).Return(nil, errors.Errorf("some error"))

	stateWriter := sherlock.NewSherlockStateWriter(suite.state, mockStateWriter)
	err := stateWriter.WriteEnvironments()
	suite.Assert().ErrorContains(err, "some error")
	mockStateWriter.AssertCalled(suite.T(), "WriteEnvironments", mock.Anything)
}

func (suite *sherlockStateWriterSuite) TestErrorWriteClusters() {
	mockStateWriter := mocks.NewStateWriter(suite.T())
	mockStateWriter.On("WriteClusters", mock.AnythingOfType("[]terra.Cluster")).Return(errors.Errorf("some error"))

	stateWriter := sherlock.NewSherlockStateWriter(suite.state, mockStateWriter)
	err := stateWriter.WriteClusters()
	suite.Assert().ErrorContains(err, "some error")
	mockStateWriter.AssertCalled(suite.T(), "WriteClusters", mock.Anything)
}

func constructFakeState(t *testing.T) terra.State {
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)
	builder := builder.NewBuilder().WithTestDefaults(t).UseCustomStateLoader(fixture.Mocks().StateLoader)
	app, err := builder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)
	return state
}
