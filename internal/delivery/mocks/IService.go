// Code generated by mockery v2.53.4. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/hinha/library-management-synapsis/internal/domain"
	mock "github.com/stretchr/testify/mock"

	protouser "github.com/hinha/library-management-synapsis/gen/api/proto/user"

	user "github.com/hinha/library-management-synapsis/internal/domain/user"
)

// IService is an autogenerated mock type for the IService type
type IService struct {
	mock.Mock
}

// GetUser provides a mock function with given fields: ctx, id
func (_m *IService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetUser")
	}

	var r0 *domain.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*domain.User, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *domain.User); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Health provides a mock function with given fields: ctx
func (_m *IService) Health(ctx context.Context) (*protouser.HealthCheckResponse, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Health")
	}

	var r0 *protouser.HealthCheckResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*protouser.HealthCheckResponse, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *protouser.HealthCheckResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*protouser.HealthCheckResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Login provides a mock function with given fields: ctx, email, password
func (_m *IService) Login(ctx context.Context, email string, password string) (string, string, error) {
	ret := _m.Called(ctx, email, password)

	if len(ret) == 0 {
		panic("no return value specified for Login")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, string, error)); ok {
		return rf(ctx, email, password)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, email, password)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) string); ok {
		r1 = rf(ctx, email, password)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(ctx, email, password)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Register provides a mock function with given fields: ctx, name, email, password, isAdmin
func (_m *IService) Register(ctx context.Context, name string, email string, password string, isAdmin bool) (*domain.User, error) {
	ret := _m.Called(ctx, name, email, password, isAdmin)

	if len(ret) == 0 {
		panic("no return value specified for Register")
	}

	var r0 *domain.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, bool) (*domain.User, error)); ok {
		return rf(ctx, name, email, password, isAdmin)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, bool) *domain.User); ok {
		r0 = rf(ctx, name, email, password, isAdmin)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, bool) error); ok {
		r1 = rf(ctx, name, email, password, isAdmin)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUser provides a mock function with given fields: ctx, id, name, email
func (_m *IService) UpdateUser(ctx context.Context, id string, name string, email string) (*domain.User, error) {
	ret := _m.Called(ctx, id, name, email)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUser")
	}

	var r0 *domain.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*domain.User, error)); ok {
		return rf(ctx, id, name, email)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *domain.User); ok {
		r0 = rf(ctx, id, name, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, id, name, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateToken provides a mock function with given fields: ctx, token
func (_m *IService) ValidateToken(ctx context.Context, token string) (*user.Claims, error) {
	ret := _m.Called(ctx, token)

	if len(ret) == 0 {
		panic("no return value specified for ValidateToken")
	}

	var r0 *user.Claims
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*user.Claims, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *user.Claims); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*user.Claims)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIService creates a new instance of IService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIService(t interface {
	mock.TestingT
	Cleanup(func())
}) *IService {
	mock := &IService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
