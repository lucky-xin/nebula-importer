// Code generated by MockGen. DO NOT EDIT.
// Source: record.go

// Package reader is a generated GoMock package.
package reader

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	source "github.com/lucky-xin/nebula-importer/pkg/source"
	spec "github.com/lucky-xin/nebula-importer/pkg/spec"
)

// MockRecordReader is a mock of RecordReader interface.
type MockRecordReader struct {
	ctrl     *gomock.Controller
	recorder *MockRecordReaderMockRecorder
}

// MockRecordReaderMockRecorder is the mock recorder for MockRecordReader.
type MockRecordReaderMockRecorder struct {
	mock *MockRecordReader
}

// NewMockRecordReader creates a new mock instance.
func NewMockRecordReader(ctrl *gomock.Controller) *MockRecordReader {
	mock := &MockRecordReader{ctrl: ctrl}
	mock.recorder = &MockRecordReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRecordReader) EXPECT() *MockRecordReaderMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockRecordReader) Read() (int, spec.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(spec.Record)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Read indicates an expected call of Read.
func (mr *MockRecordReaderMockRecorder) Read() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockRecordReader)(nil).Read))
}

// Size mocks base method.
func (m *MockRecordReader) Size() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Size indicates an expected call of Size.
func (mr *MockRecordReaderMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockRecordReader)(nil).Size))
}

// Source mocks base method.
func (m *MockRecordReader) Source() source.Source {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Source")
	ret0, _ := ret[0].(source.Source)
	return ret0
}

// Source indicates an expected call of Source.
func (mr *MockRecordReaderMockRecorder) Source() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Source", reflect.TypeOf((*MockRecordReader)(nil).Source))
}
