// Code generated by mockery v2.37.1. DO NOT EDIT.

package e2e

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	metadata "google.golang.org/grpc/metadata"

	proto "github.com/dhiaayachi/consul/internal/tools/protoc-gen-grpc-clone/e2e/proto"
)

// Simple_FlowClient is an autogenerated mock type for the Simple_FlowClient type
type Simple_FlowClient struct {
	mock.Mock
}

type Simple_FlowClient_Expecter struct {
	mock *mock.Mock
}

func (_m *Simple_FlowClient) EXPECT() *Simple_FlowClient_Expecter {
	return &Simple_FlowClient_Expecter{mock: &_m.Mock}
}

// CloseSend provides a mock function with given fields:
func (_m *Simple_FlowClient) CloseSend() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Simple_FlowClient_CloseSend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CloseSend'
type Simple_FlowClient_CloseSend_Call struct {
	*mock.Call
}

// CloseSend is a helper method to define mock.On call
func (_e *Simple_FlowClient_Expecter) CloseSend() *Simple_FlowClient_CloseSend_Call {
	return &Simple_FlowClient_CloseSend_Call{Call: _e.mock.On("CloseSend")}
}

func (_c *Simple_FlowClient_CloseSend_Call) Run(run func()) *Simple_FlowClient_CloseSend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Simple_FlowClient_CloseSend_Call) Return(_a0 error) *Simple_FlowClient_CloseSend_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Simple_FlowClient_CloseSend_Call) RunAndReturn(run func() error) *Simple_FlowClient_CloseSend_Call {
	_c.Call.Return(run)
	return _c
}

// Context provides a mock function with given fields:
func (_m *Simple_FlowClient) Context() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// Simple_FlowClient_Context_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Context'
type Simple_FlowClient_Context_Call struct {
	*mock.Call
}

// Context is a helper method to define mock.On call
func (_e *Simple_FlowClient_Expecter) Context() *Simple_FlowClient_Context_Call {
	return &Simple_FlowClient_Context_Call{Call: _e.mock.On("Context")}
}

func (_c *Simple_FlowClient_Context_Call) Run(run func()) *Simple_FlowClient_Context_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Simple_FlowClient_Context_Call) Return(_a0 context.Context) *Simple_FlowClient_Context_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Simple_FlowClient_Context_Call) RunAndReturn(run func() context.Context) *Simple_FlowClient_Context_Call {
	_c.Call.Return(run)
	return _c
}

// Header provides a mock function with given fields:
func (_m *Simple_FlowClient) Header() (metadata.MD, error) {
	ret := _m.Called()

	var r0 metadata.MD
	var r1 error
	if rf, ok := ret.Get(0).(func() (metadata.MD, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Simple_FlowClient_Header_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Header'
type Simple_FlowClient_Header_Call struct {
	*mock.Call
}

// Header is a helper method to define mock.On call
func (_e *Simple_FlowClient_Expecter) Header() *Simple_FlowClient_Header_Call {
	return &Simple_FlowClient_Header_Call{Call: _e.mock.On("Header")}
}

func (_c *Simple_FlowClient_Header_Call) Run(run func()) *Simple_FlowClient_Header_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Simple_FlowClient_Header_Call) Return(_a0 metadata.MD, _a1 error) *Simple_FlowClient_Header_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Simple_FlowClient_Header_Call) RunAndReturn(run func() (metadata.MD, error)) *Simple_FlowClient_Header_Call {
	_c.Call.Return(run)
	return _c
}

// Recv provides a mock function with given fields:
func (_m *Simple_FlowClient) Recv() (*proto.Resp, error) {
	ret := _m.Called()

	var r0 *proto.Resp
	var r1 error
	if rf, ok := ret.Get(0).(func() (*proto.Resp, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *proto.Resp); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*proto.Resp)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Simple_FlowClient_Recv_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Recv'
type Simple_FlowClient_Recv_Call struct {
	*mock.Call
}

// Recv is a helper method to define mock.On call
func (_e *Simple_FlowClient_Expecter) Recv() *Simple_FlowClient_Recv_Call {
	return &Simple_FlowClient_Recv_Call{Call: _e.mock.On("Recv")}
}

func (_c *Simple_FlowClient_Recv_Call) Run(run func()) *Simple_FlowClient_Recv_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Simple_FlowClient_Recv_Call) Return(_a0 *proto.Resp, _a1 error) *Simple_FlowClient_Recv_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Simple_FlowClient_Recv_Call) RunAndReturn(run func() (*proto.Resp, error)) *Simple_FlowClient_Recv_Call {
	_c.Call.Return(run)
	return _c
}

// RecvMsg provides a mock function with given fields: m
func (_m *Simple_FlowClient) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Simple_FlowClient_RecvMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecvMsg'
type Simple_FlowClient_RecvMsg_Call struct {
	*mock.Call
}

// RecvMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *Simple_FlowClient_Expecter) RecvMsg(m interface{}) *Simple_FlowClient_RecvMsg_Call {
	return &Simple_FlowClient_RecvMsg_Call{Call: _e.mock.On("RecvMsg", m)}
}

func (_c *Simple_FlowClient_RecvMsg_Call) Run(run func(m interface{})) *Simple_FlowClient_RecvMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Simple_FlowClient_RecvMsg_Call) Return(_a0 error) *Simple_FlowClient_RecvMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Simple_FlowClient_RecvMsg_Call) RunAndReturn(run func(interface{}) error) *Simple_FlowClient_RecvMsg_Call {
	_c.Call.Return(run)
	return _c
}

// SendMsg provides a mock function with given fields: m
func (_m *Simple_FlowClient) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Simple_FlowClient_SendMsg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendMsg'
type Simple_FlowClient_SendMsg_Call struct {
	*mock.Call
}

// SendMsg is a helper method to define mock.On call
//   - m interface{}
func (_e *Simple_FlowClient_Expecter) SendMsg(m interface{}) *Simple_FlowClient_SendMsg_Call {
	return &Simple_FlowClient_SendMsg_Call{Call: _e.mock.On("SendMsg", m)}
}

func (_c *Simple_FlowClient_SendMsg_Call) Run(run func(m interface{})) *Simple_FlowClient_SendMsg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *Simple_FlowClient_SendMsg_Call) Return(_a0 error) *Simple_FlowClient_SendMsg_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Simple_FlowClient_SendMsg_Call) RunAndReturn(run func(interface{}) error) *Simple_FlowClient_SendMsg_Call {
	_c.Call.Return(run)
	return _c
}

// Trailer provides a mock function with given fields:
func (_m *Simple_FlowClient) Trailer() metadata.MD {
	ret := _m.Called()

	var r0 metadata.MD
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	return r0
}

// Simple_FlowClient_Trailer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Trailer'
type Simple_FlowClient_Trailer_Call struct {
	*mock.Call
}

// Trailer is a helper method to define mock.On call
func (_e *Simple_FlowClient_Expecter) Trailer() *Simple_FlowClient_Trailer_Call {
	return &Simple_FlowClient_Trailer_Call{Call: _e.mock.On("Trailer")}
}

func (_c *Simple_FlowClient_Trailer_Call) Run(run func()) *Simple_FlowClient_Trailer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Simple_FlowClient_Trailer_Call) Return(_a0 metadata.MD) *Simple_FlowClient_Trailer_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Simple_FlowClient_Trailer_Call) RunAndReturn(run func() metadata.MD) *Simple_FlowClient_Trailer_Call {
	_c.Call.Return(run)
	return _c
}

// NewSimple_FlowClient creates a new instance of Simple_FlowClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSimple_FlowClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Simple_FlowClient {
	mock := &Simple_FlowClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
