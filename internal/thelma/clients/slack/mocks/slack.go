// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Slack is an autogenerated mock type for the Slack type
type Slack struct {
	mock.Mock
}

type Slack_Expecter struct {
	mock *mock.Mock
}

func (_m *Slack) EXPECT() *Slack_Expecter {
	return &Slack_Expecter{mock: &_m.Mock}
}

// SendDevopsAlert provides a mock function with given fields: title, text, ok
func (_m *Slack) SendDevopsAlert(title string, text string, ok bool) error {
	ret := _m.Called(title, text, ok)

	if len(ret) == 0 {
		panic("no return value specified for SendDevopsAlert")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, bool) error); ok {
		r0 = rf(title, text, ok)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Slack_SendDevopsAlert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendDevopsAlert'
type Slack_SendDevopsAlert_Call struct {
	*mock.Call
}

// SendDevopsAlert is a helper method to define mock.On call
//   - title string
//   - text string
//   - ok bool
func (_e *Slack_Expecter) SendDevopsAlert(title interface{}, text interface{}, ok interface{}) *Slack_SendDevopsAlert_Call {
	return &Slack_SendDevopsAlert_Call{Call: _e.mock.On("SendDevopsAlert", title, text, ok)}
}

func (_c *Slack_SendDevopsAlert_Call) Run(run func(title string, text string, ok bool)) *Slack_SendDevopsAlert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(bool))
	})
	return _c
}

func (_c *Slack_SendDevopsAlert_Call) Return(_a0 error) *Slack_SendDevopsAlert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Slack_SendDevopsAlert_Call) RunAndReturn(run func(string, string, bool) error) *Slack_SendDevopsAlert_Call {
	_c.Call.Return(run)
	return _c
}

// SendDirectMessage provides a mock function with given fields: email, markdown
func (_m *Slack) SendDirectMessage(email string, markdown string) error {
	ret := _m.Called(email, markdown)

	if len(ret) == 0 {
		panic("no return value specified for SendDirectMessage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(email, markdown)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Slack_SendDirectMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendDirectMessage'
type Slack_SendDirectMessage_Call struct {
	*mock.Call
}

// SendDirectMessage is a helper method to define mock.On call
//   - email string
//   - markdown string
func (_e *Slack_Expecter) SendDirectMessage(email interface{}, markdown interface{}) *Slack_SendDirectMessage_Call {
	return &Slack_SendDirectMessage_Call{Call: _e.mock.On("SendDirectMessage", email, markdown)}
}

func (_c *Slack_SendDirectMessage_Call) Run(run func(email string, markdown string)) *Slack_SendDirectMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Slack_SendDirectMessage_Call) Return(_a0 error) *Slack_SendDirectMessage_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Slack_SendDirectMessage_Call) RunAndReturn(run func(string, string) error) *Slack_SendDirectMessage_Call {
	_c.Call.Return(run)
	return _c
}

// NewSlack creates a new instance of Slack. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSlack(t interface {
	mock.TestingT
	Cleanup(func())
}) *Slack {
	mock := &Slack{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}