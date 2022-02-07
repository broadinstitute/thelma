package app

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/paths"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/mock"
)

type MockThelmaApp struct {
	mock.Mock
}

func NewMockApp() ThelmaApp {
	return &MockThelmaApp{}
}

func (m *MockThelmaApp) Config() config.Config {
	return m.Called().Get(0).(config.Config)
}

func (m *MockThelmaApp) Paths() paths.Paths {
	return m.Called().Get(0).(paths.Paths)
}

func (m *MockThelmaApp) Scratch() scratch.Scratch {
	return m.Called().Get(0).(scratch.Scratch)
}

func (m *MockThelmaApp) ShellRunner() shell.Runner {
	return m.Called().Get(0).(shell.Runner)
}

func (m *MockThelmaApp) Close() error {
	return m.Called().Error(0)
}
