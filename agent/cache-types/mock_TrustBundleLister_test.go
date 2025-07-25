// Code generated by mockery v2.15.0. DO NOT EDIT.

package cachetype

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	pbpeering "github.com/dhiaayachi/consul/proto/private/pbpeering"
)

// MockTrustBundleLister is an autogenerated mock type for the TrustBundleLister type
type MockTrustBundleLister struct {
	mock.Mock
}

// TrustBundleListByService provides a mock function with given fields: ctx, in, opts
func (_m *MockTrustBundleLister) TrustBundleListByService(ctx context.Context, in *pbpeering.TrustBundleListByServiceRequest, opts ...grpc.CallOption) (*pbpeering.TrustBundleListByServiceResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *pbpeering.TrustBundleListByServiceResponse
	if rf, ok := ret.Get(0).(func(context.Context, *pbpeering.TrustBundleListByServiceRequest, ...grpc.CallOption) *pbpeering.TrustBundleListByServiceResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pbpeering.TrustBundleListByServiceResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pbpeering.TrustBundleListByServiceRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockTrustBundleLister interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockTrustBundleLister creates a new instance of MockTrustBundleLister. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockTrustBundleLister(t mockConstructorTestingTNewMockTrustBundleLister) *MockTrustBundleLister {
	mock := &MockTrustBundleLister{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
