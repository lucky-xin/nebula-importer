// Code generated by MockGen. DO NOT EDIT.
// Source: importer.go

// Package importer is a generated GoMock package.
package importer

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	spec "github.com/lucky-xin/nebula-importer/pkg/spec"
)

// MockImporter is a mock of Importer interface.
type MockImporter struct {
	ctrl     *gomock.Controller
	recorder *MockImporterMockRecorder
}

// MockImporterMockRecorder is the mock recorder for MockImporter.
type MockImporterMockRecorder struct {
	mock *MockImporter
}

// NewMockImporter creates a new mock instance.
func NewMockImporter(ctrl *gomock.Controller) *MockImporter {
	mock := &MockImporter{ctrl: ctrl}
	mock.recorder = &MockImporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockImporter) EXPECT() *MockImporterMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockImporter) Add(delta int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", delta)
}

// Add indicates an expected call of Add.
func (mr *MockImporterMockRecorder) Add(delta interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockImporter)(nil).Add), delta)
}

// Done mocks base method.
func (m *MockImporter) Done() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Done")
}

// Done indicates an expected call of Done.
func (mr *MockImporterMockRecorder) Done() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Done", reflect.TypeOf((*MockImporter)(nil).Done))
}

// Import mocks base method.
func (m *MockImporter) Import(records ...spec.Record) (*ImportResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range records {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Import", varargs...)
	ret0, _ := ret[0].(*ImportResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Import indicates an expected call of Import.
func (mr *MockImporterMockRecorder) Import(records ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Import", reflect.TypeOf((*MockImporter)(nil).Import), records...)
}

// Wait mocks base method.
func (m *MockImporter) Wait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Wait")
}

// Wait indicates an expected call of Wait.
func (mr *MockImporterMockRecorder) Wait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockImporter)(nil).Wait))
}
