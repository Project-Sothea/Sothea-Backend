// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	entities "github.com/jieqiboh/sothea_backend/entities"
	mock "github.com/stretchr/testify/mock"
)

// PatientRepository is an autogenerated mock type for the PatientRepository type
type PatientRepository struct {
	mock.Mock
}

// CreatePatient provides a mock function with given fields: ctx, admin
func (_m *PatientRepository) CreatePatient(ctx context.Context, admin *entities.Admin) (int32, error) {
	ret := _m.Called(ctx, admin)

	if len(ret) == 0 {
		panic("no return value specified for CreatePatient")
	}

	var r0 int32
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *entities.Admin) (int32, error)); ok {
		return rf(ctx, admin)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *entities.Admin) int32); ok {
		r0 = rf(ctx, admin)
	} else {
		r0 = ret.Get(0).(int32)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *entities.Admin) error); ok {
		r1 = rf(ctx, admin)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePatientVisit provides a mock function with given fields: ctx, id, admin
func (_m *PatientRepository) CreatePatientVisit(ctx context.Context, id int32, admin *entities.Admin) (int32, error) {
	ret := _m.Called(ctx, id, admin)

	if len(ret) == 0 {
		panic("no return value specified for CreatePatientVisit")
	}

	var r0 int32
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int32, *entities.Admin) (int32, error)); ok {
		return rf(ctx, id, admin)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int32, *entities.Admin) int32); ok {
		r0 = rf(ctx, id, admin)
	} else {
		r0 = ret.Get(0).(int32)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int32, *entities.Admin) error); ok {
		r1 = rf(ctx, id, admin)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeletePatientVisit provides a mock function with given fields: ctx, id, vid
func (_m *PatientRepository) DeletePatientVisit(ctx context.Context, id int32, vid int32) error {
	ret := _m.Called(ctx, id, vid)

	if len(ret) == 0 {
		panic("no return value specified for DeletePatientVisit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int32, int32) error); ok {
		r0 = rf(ctx, id, vid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExportDatabaseToCSV provides a mock function with given fields: ctx
func (_m *PatientRepository) ExportDatabaseToCSV(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ExportDatabaseToCSV")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllAdmin provides a mock function with given fields: ctx
func (_m *PatientRepository) GetAllAdmin(ctx context.Context) ([]entities.PartAdmin, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAllAdmin")
	}

	var r0 []entities.PartAdmin
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]entities.PartAdmin, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []entities.PartAdmin); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.PartAdmin)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPatientVisit provides a mock function with given fields: ctx, id, vid
func (_m *PatientRepository) GetPatientVisit(ctx context.Context, id int32, vid int32) (*entities.Patient, error) {
	ret := _m.Called(ctx, id, vid)

	if len(ret) == 0 {
		panic("no return value specified for GetPatientVisit")
	}

	var r0 *entities.Patient
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int32, int32) (*entities.Patient, error)); ok {
		return rf(ctx, id, vid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int32, int32) *entities.Patient); ok {
		r0 = rf(ctx, id, vid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entities.Patient)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int32, int32) error); ok {
		r1 = rf(ctx, id, vid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchPatients provides a mock function with given fields: ctx, search
func (_m *PatientRepository) SearchPatients(ctx context.Context, search string) ([]entities.PartAdmin, error) {
	ret := _m.Called(ctx, search)

	if len(ret) == 0 {
		panic("no return value specified for SearchPatients")
	}

	var r0 []entities.PartAdmin
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]entities.PartAdmin, error)); ok {
		return rf(ctx, search)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []entities.PartAdmin); ok {
		r0 = rf(ctx, search)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.PartAdmin)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, search)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePatientVisit provides a mock function with given fields: ctx, id, vid, patient
func (_m *PatientRepository) UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *entities.Patient) error {
	ret := _m.Called(ctx, id, vid, patient)

	if len(ret) == 0 {
		panic("no return value specified for UpdatePatientVisit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int32, int32, *entities.Patient) error); ok {
		r0 = rf(ctx, id, vid, patient)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewPatientRepository creates a new instance of PatientRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPatientRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *PatientRepository {
	mock := &PatientRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
