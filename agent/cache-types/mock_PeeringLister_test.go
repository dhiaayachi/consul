// Code generated by mockery v2.15.0. DO NOT EDIT.

package cachetype

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	pbpeering "github.com/dhiaayachi/consul/proto/private/pbpeering"
)

// MockPeeringLister is an autogenerated mock type for the PeeringLister type
type MockPeeringLister struct {
	mock.Mock
}

// PeeringList provides a mock function with given fields: ctx, in, opts
func (_m *MockPeeringLister) PeeringList(ctx context.Context, in *pbpeering.PeeringListRequest, opts ...grpc.CallOption) (*pbpeering.PeeringListResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *pbpeering.PeeringListResponse
	if rf, ok := ret.Get(0).(func(context.Context, *pbpeering.PeeringListRequest, ...grpc.CallOption) *pbpeering.PeeringListResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pbpeering.PeeringListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pbpeering.PeeringListRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockPeeringLister interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockPeeringLister creates a new instance of MockPeeringLister. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockPeeringLister(t mockConstructorTestingTNewMockPeeringLister) *MockPeeringLister {
	mock := &MockPeeringLister{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
