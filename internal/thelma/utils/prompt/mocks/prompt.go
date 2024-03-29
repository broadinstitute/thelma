// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	prompt "github.com/broadinstitute/thelma/internal/thelma/utils/prompt"
	mock "github.com/stretchr/testify/mock"
)

// Prompt is an autogenerated mock type for the Prompt type
type Prompt struct {
	mock.Mock
}

type Prompt_Expecter struct {
	mock *mock.Mock
}

func (_m *Prompt) EXPECT() *Prompt_Expecter {
	return &Prompt_Expecter{mock: &_m.Mock}
}

// Confirm provides a mock function with given fields: message, options
func (_m *Prompt) Confirm(message string, options ...func(*prompt.ConfirmOptions)) (bool, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...func(*prompt.ConfirmOptions)) (bool, error)); ok {
		return rf(message, options...)
	}
	if rf, ok := ret.Get(0).(func(string, ...func(*prompt.ConfirmOptions)) bool); ok {
		r0 = rf(message, options...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string, ...func(*prompt.ConfirmOptions)) error); ok {
		r1 = rf(message, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Prompt_Confirm_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Confirm'
type Prompt_Confirm_Call struct {
	*mock.Call
}

// Confirm is a helper method to define mock.On call
//   - message string
//   - options ...func(*prompt.ConfirmOptions)
func (_e *Prompt_Expecter) Confirm(message interface{}, options ...interface{}) *Prompt_Confirm_Call {
	return &Prompt_Confirm_Call{Call: _e.mock.On("Confirm",
		append([]interface{}{message}, options...)...)}
}

func (_c *Prompt_Confirm_Call) Run(run func(message string, options ...func(*prompt.ConfirmOptions))) *Prompt_Confirm_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*prompt.ConfirmOptions), len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(func(*prompt.ConfirmOptions))
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Prompt_Confirm_Call) Return(_a0 bool, _a1 error) *Prompt_Confirm_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Prompt_Confirm_Call) RunAndReturn(run func(string, ...func(*prompt.ConfirmOptions)) (bool, error)) *Prompt_Confirm_Call {
	_c.Call.Return(run)
	return _c
}

// Newline provides a mock function with given fields: count
func (_m *Prompt) Newline(count ...int) error {
	_va := make([]interface{}, len(count))
	for _i := range count {
		_va[_i] = count[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(...int) error); ok {
		r0 = rf(count...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Prompt_Newline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Newline'
type Prompt_Newline_Call struct {
	*mock.Call
}

// Newline is a helper method to define mock.On call
//   - count ...int
func (_e *Prompt_Expecter) Newline(count ...interface{}) *Prompt_Newline_Call {
	return &Prompt_Newline_Call{Call: _e.mock.On("Newline",
		append([]interface{}{}, count...)...)}
}

func (_c *Prompt_Newline_Call) Run(run func(count ...int)) *Prompt_Newline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]int, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(int)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *Prompt_Newline_Call) Return(_a0 error) *Prompt_Newline_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Prompt_Newline_Call) RunAndReturn(run func(...int) error) *Prompt_Newline_Call {
	_c.Call.Return(run)
	return _c
}

// Print provides a mock function with given fields: text, options
func (_m *Prompt) Print(text string, options ...func(*prompt.PrintOptions)) error {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, text)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, ...func(*prompt.PrintOptions)) error); ok {
		r0 = rf(text, options...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Prompt_Print_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Print'
type Prompt_Print_Call struct {
	*mock.Call
}

// Print is a helper method to define mock.On call
//   - text string
//   - options ...func(*prompt.PrintOptions)
func (_e *Prompt_Expecter) Print(text interface{}, options ...interface{}) *Prompt_Print_Call {
	return &Prompt_Print_Call{Call: _e.mock.On("Print",
		append([]interface{}{text}, options...)...)}
}

func (_c *Prompt_Print_Call) Run(run func(text string, options ...func(*prompt.PrintOptions))) *Prompt_Print_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*prompt.PrintOptions), len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(func(*prompt.PrintOptions))
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Prompt_Print_Call) Return(_a0 error) *Prompt_Print_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Prompt_Print_Call) RunAndReturn(run func(string, ...func(*prompt.PrintOptions)) error) *Prompt_Print_Call {
	_c.Call.Return(run)
	return _c
}

// NewPrompt creates a new instance of Prompt. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPrompt(t interface {
	mock.TestingT
	Cleanup(func())
}) *Prompt {
	mock := &Prompt{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
