// Code generated by mockery v2.15.0. DO NOT EDIT.

package connectca

import (
	acl "github.com/dhiaayachi/consul/acl"
	mock "github.com/stretchr/testify/mock"

	structs "github.com/dhiaayachi/consul/agent/structs"

	x509 "crypto/x509"
)

// MockCAManager is an autogenerated mock type for the CAManager type
type MockCAManager struct {
	mock.Mock
}

// AuthorizeAndSignCertificate provides a mock function with given fields: csr, authz
func (_m *MockCAManager) AuthorizeAndSignCertificate(csr *x509.CertificateRequest, authz acl.Authorizer) (*structs.IssuedCert, error) {
	ret := _m.Called(csr, authz)

	var r0 *structs.IssuedCert
	if rf, ok := ret.Get(0).(func(*x509.CertificateRequest, acl.Authorizer) *structs.IssuedCert); ok {
		r0 = rf(csr, authz)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*structs.IssuedCert)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*x509.CertificateRequest, acl.Authorizer) error); ok {
		r1 = rf(csr, authz)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockCAManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockCAManager creates a new instance of MockCAManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockCAManager(t mockConstructorTestingTNewMockCAManager) *MockCAManager {
	mock := &MockCAManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
