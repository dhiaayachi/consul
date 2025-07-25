// Code generated by mockery v2.20.0. DO NOT EDIT.

package v1compat

import (
	context "context"

	acl "github.com/dhiaayachi/consul/acl"

	controller "github.com/dhiaayachi/consul/internal/controller"

	mock "github.com/stretchr/testify/mock"

	structs "github.com/dhiaayachi/consul/agent/structs"
)

// MockAggregatedConfig is an autogenerated mock type for the AggregatedConfig type
type MockAggregatedConfig struct {
	mock.Mock
}

type MockAggregatedConfig_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAggregatedConfig) EXPECT() *MockAggregatedConfig_Expecter {
	return &MockAggregatedConfig_Expecter{mock: &_m.Mock}
}

// DeleteExportedServicesConfigEntry provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockAggregatedConfig) DeleteExportedServicesConfigEntry(_a0 context.Context, _a1 string, _a2 *acl.EnterpriseMeta) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *acl.EnterpriseMeta) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteExportedServicesConfigEntry'
type MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call struct {
	*mock.Call
}

// DeleteExportedServicesConfigEntry is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 *acl.EnterpriseMeta
func (_e *MockAggregatedConfig_Expecter) DeleteExportedServicesConfigEntry(_a0 interface{}, _a1 interface{}, _a2 interface{}) *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call {
	return &MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call{Call: _e.mock.On("DeleteExportedServicesConfigEntry", _a0, _a1, _a2)}
}

func (_c *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call) Run(run func(_a0 context.Context, _a1 string, _a2 *acl.EnterpriseMeta)) *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*acl.EnterpriseMeta))
	})
	return _c
}

func (_c *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call) Return(_a0 error) *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call) RunAndReturn(run func(context.Context, string, *acl.EnterpriseMeta) error) *MockAggregatedConfig_DeleteExportedServicesConfigEntry_Call {
	_c.Call.Return(run)
	return _c
}

// EventChannel provides a mock function with given fields:
func (_m *MockAggregatedConfig) EventChannel() chan controller.Event {
	ret := _m.Called()

	var r0 chan controller.Event
	if rf, ok := ret.Get(0).(func() chan controller.Event); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan controller.Event)
		}
	}

	return r0
}

// MockAggregatedConfig_EventChannel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EventChannel'
type MockAggregatedConfig_EventChannel_Call struct {
	*mock.Call
}

// EventChannel is a helper method to define mock.On call
func (_e *MockAggregatedConfig_Expecter) EventChannel() *MockAggregatedConfig_EventChannel_Call {
	return &MockAggregatedConfig_EventChannel_Call{Call: _e.mock.On("EventChannel")}
}

func (_c *MockAggregatedConfig_EventChannel_Call) Run(run func()) *MockAggregatedConfig_EventChannel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockAggregatedConfig_EventChannel_Call) Return(_a0 chan controller.Event) *MockAggregatedConfig_EventChannel_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAggregatedConfig_EventChannel_Call) RunAndReturn(run func() chan controller.Event) *MockAggregatedConfig_EventChannel_Call {
	_c.Call.Return(run)
	return _c
}

// GetExportedServicesConfigEntry provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockAggregatedConfig) GetExportedServicesConfigEntry(_a0 context.Context, _a1 string, _a2 *acl.EnterpriseMeta) (*structs.ExportedServicesConfigEntry, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *structs.ExportedServicesConfigEntry
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *acl.EnterpriseMeta) (*structs.ExportedServicesConfigEntry, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *acl.EnterpriseMeta) *structs.ExportedServicesConfigEntry); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*structs.ExportedServicesConfigEntry)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *acl.EnterpriseMeta) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAggregatedConfig_GetExportedServicesConfigEntry_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExportedServicesConfigEntry'
type MockAggregatedConfig_GetExportedServicesConfigEntry_Call struct {
	*mock.Call
}

// GetExportedServicesConfigEntry is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 *acl.EnterpriseMeta
func (_e *MockAggregatedConfig_Expecter) GetExportedServicesConfigEntry(_a0 interface{}, _a1 interface{}, _a2 interface{}) *MockAggregatedConfig_GetExportedServicesConfigEntry_Call {
	return &MockAggregatedConfig_GetExportedServicesConfigEntry_Call{Call: _e.mock.On("GetExportedServicesConfigEntry", _a0, _a1, _a2)}
}

func (_c *MockAggregatedConfig_GetExportedServicesConfigEntry_Call) Run(run func(_a0 context.Context, _a1 string, _a2 *acl.EnterpriseMeta)) *MockAggregatedConfig_GetExportedServicesConfigEntry_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*acl.EnterpriseMeta))
	})
	return _c
}

func (_c *MockAggregatedConfig_GetExportedServicesConfigEntry_Call) Return(_a0 *structs.ExportedServicesConfigEntry, _a1 error) *MockAggregatedConfig_GetExportedServicesConfigEntry_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAggregatedConfig_GetExportedServicesConfigEntry_Call) RunAndReturn(run func(context.Context, string, *acl.EnterpriseMeta) (*structs.ExportedServicesConfigEntry, error)) *MockAggregatedConfig_GetExportedServicesConfigEntry_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: _a0
func (_m *MockAggregatedConfig) Start(_a0 context.Context) {
	_m.Called(_a0)
}

// MockAggregatedConfig_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockAggregatedConfig_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockAggregatedConfig_Expecter) Start(_a0 interface{}) *MockAggregatedConfig_Start_Call {
	return &MockAggregatedConfig_Start_Call{Call: _e.mock.On("Start", _a0)}
}

func (_c *MockAggregatedConfig_Start_Call) Run(run func(_a0 context.Context)) *MockAggregatedConfig_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockAggregatedConfig_Start_Call) Return() *MockAggregatedConfig_Start_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockAggregatedConfig_Start_Call) RunAndReturn(run func(context.Context)) *MockAggregatedConfig_Start_Call {
	_c.Call.Return(run)
	return _c
}

// WriteExportedServicesConfigEntry provides a mock function with given fields: _a0, _a1
func (_m *MockAggregatedConfig) WriteExportedServicesConfigEntry(_a0 context.Context, _a1 *structs.ExportedServicesConfigEntry) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *structs.ExportedServicesConfigEntry) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAggregatedConfig_WriteExportedServicesConfigEntry_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteExportedServicesConfigEntry'
type MockAggregatedConfig_WriteExportedServicesConfigEntry_Call struct {
	*mock.Call
}

// WriteExportedServicesConfigEntry is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *structs.ExportedServicesConfigEntry
func (_e *MockAggregatedConfig_Expecter) WriteExportedServicesConfigEntry(_a0 interface{}, _a1 interface{}) *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call {
	return &MockAggregatedConfig_WriteExportedServicesConfigEntry_Call{Call: _e.mock.On("WriteExportedServicesConfigEntry", _a0, _a1)}
}

func (_c *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call) Run(run func(_a0 context.Context, _a1 *structs.ExportedServicesConfigEntry)) *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*structs.ExportedServicesConfigEntry))
	})
	return _c
}

func (_c *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call) Return(_a0 error) *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call) RunAndReturn(run func(context.Context, *structs.ExportedServicesConfigEntry) error) *MockAggregatedConfig_WriteExportedServicesConfigEntry_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewMockAggregatedConfig interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockAggregatedConfig creates a new instance of MockAggregatedConfig. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockAggregatedConfig(t mockConstructorTestingTNewMockAggregatedConfig) *MockAggregatedConfig {
	mock := &MockAggregatedConfig{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
