// Code generated by mockery v2.42.2. DO NOT EDIT.

package authentication

import (
	context "context"

	logger "github.com/neticdk-k8s/k8s-inventory-cli/internal/logger"
	mock "github.com/stretchr/testify/mock"
)

// MockAuthentication is an autogenerated mock type for the Authentication type
type MockAuthentication struct {
	mock.Mock
}

type MockAuthentication_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAuthentication) EXPECT() *MockAuthentication_Expecter {
	return &MockAuthentication_Expecter{mock: &_m.Mock}
}

// Authenticate provides a mock function with given fields: ctx, in
func (_m *MockAuthentication) Authenticate(ctx context.Context, in AuthenticateInput) (*AuthResult, error) {
	ret := _m.Called(ctx, in)

	if len(ret) == 0 {
		panic("no return value specified for Authenticate")
	}

	var r0 *AuthResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, AuthenticateInput) (*AuthResult, error)); ok {
		return rf(ctx, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, AuthenticateInput) *AuthResult); ok {
		r0 = rf(ctx, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*AuthResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, AuthenticateInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAuthentication_Authenticate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Authenticate'
type MockAuthentication_Authenticate_Call struct {
	*mock.Call
}

// Authenticate is a helper method to define mock.On call
//   - ctx context.Context
//   - in AuthenticateInput
func (_e *MockAuthentication_Expecter) Authenticate(ctx interface{}, in interface{}) *MockAuthentication_Authenticate_Call {
	return &MockAuthentication_Authenticate_Call{Call: _e.mock.On("Authenticate", ctx, in)}
}

func (_c *MockAuthentication_Authenticate_Call) Run(run func(ctx context.Context, in AuthenticateInput)) *MockAuthentication_Authenticate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(AuthenticateInput))
	})
	return _c
}

func (_c *MockAuthentication_Authenticate_Call) Return(_a0 *AuthResult, _a1 error) *MockAuthentication_Authenticate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAuthentication_Authenticate_Call) RunAndReturn(run func(context.Context, AuthenticateInput) (*AuthResult, error)) *MockAuthentication_Authenticate_Call {
	_c.Call.Return(run)
	return _c
}

// Logout provides a mock function with given fields: ctx, in
func (_m *MockAuthentication) Logout(ctx context.Context, in AuthenticateLogoutInput) error {
	ret := _m.Called(ctx, in)

	if len(ret) == 0 {
		panic("no return value specified for Logout")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, AuthenticateLogoutInput) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAuthentication_Logout_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Logout'
type MockAuthentication_Logout_Call struct {
	*mock.Call
}

// Logout is a helper method to define mock.On call
//   - ctx context.Context
//   - in AuthenticateLogoutInput
func (_e *MockAuthentication_Expecter) Logout(ctx interface{}, in interface{}) *MockAuthentication_Logout_Call {
	return &MockAuthentication_Logout_Call{Call: _e.mock.On("Logout", ctx, in)}
}

func (_c *MockAuthentication_Logout_Call) Run(run func(ctx context.Context, in AuthenticateLogoutInput)) *MockAuthentication_Logout_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(AuthenticateLogoutInput))
	})
	return _c
}

func (_c *MockAuthentication_Logout_Call) Return(_a0 error) *MockAuthentication_Logout_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAuthentication_Logout_Call) RunAndReturn(run func(context.Context, AuthenticateLogoutInput) error) *MockAuthentication_Logout_Call {
	_c.Call.Return(run)
	return _c
}

// SetLogger provides a mock function with given fields: _a0
func (_m *MockAuthentication) SetLogger(_a0 logger.Logger) {
	_m.Called(_a0)
}

// MockAuthentication_SetLogger_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetLogger'
type MockAuthentication_SetLogger_Call struct {
	*mock.Call
}

// SetLogger is a helper method to define mock.On call
//   - _a0 logger.Logger
func (_e *MockAuthentication_Expecter) SetLogger(_a0 interface{}) *MockAuthentication_SetLogger_Call {
	return &MockAuthentication_SetLogger_Call{Call: _e.mock.On("SetLogger", _a0)}
}

func (_c *MockAuthentication_SetLogger_Call) Run(run func(_a0 logger.Logger)) *MockAuthentication_SetLogger_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(logger.Logger))
	})
	return _c
}

func (_c *MockAuthentication_SetLogger_Call) Return() *MockAuthentication_SetLogger_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockAuthentication_SetLogger_Call) RunAndReturn(run func(logger.Logger)) *MockAuthentication_SetLogger_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAuthentication creates a new instance of MockAuthentication. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAuthentication(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuthentication {
	mock := &MockAuthentication{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
