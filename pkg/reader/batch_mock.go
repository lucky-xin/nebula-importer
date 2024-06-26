// Code generated by MockGen. DO NOT EDIT.
// Source: batch.go

// Package reader is a generated GoMock package.
package reader

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	source "github.com/lucky-xin/nebula-importer/pkg/source"
	spec "github.com/lucky-xin/nebula-importer/pkg/spec"
)

// MockBatchRecordReader is a mock of BatchRecordReader interface.
type MockBatchRecordReader struct {
	ctrl     *gomock.Controller
	recorder *MockBatchRecordReaderMockRecorder
}

// MockBatchRecordReaderMockRecorder is the mock recorder for MockBatchRecordReader.
type MockBatchRecordReaderMockRecorder struct {
	mock *MockBatchRecordReader
}

// NewMockBatchRecordReader creates a new mock instance.
func NewMockBatchRecordReader(ctrl *gomock.Controller) *MockBatchRecordReader {
	mock := &MockBatchRecordReader{ctrl: ctrl}
	mock.recorder = &MockBatchRecordReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBatchRecordReader) EXPECT() *MockBatchRecordReaderMockRecorder {
	return m.recorder
}

// ReadBatch mocks base method.
func (m *MockBatchRecordReader) ReadBatch() (int, spec.Records, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadBatch")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(spec.Records)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ReadBatch indicates an expected call of ReadBatch.
func (mr *MockBatchRecordReaderMockRecorder) ReadBatch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadBatch", reflect.TypeOf((*MockBatchRecordReader)(nil).ReadBatch))
}

// Size mocks base method.
func (m *MockBatchRecordReader) Size() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Size indicates an expected call of Size.
func (mr *MockBatchRecordReaderMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockBatchRecordReader)(nil).Size))
}

// Source mocks base method.
func (m *MockBatchRecordReader) Source() source.Source {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Source")
	ret0, _ := ret[0].(source.Source)
	return ret0
}

// Source indicates an expected call of Source.
func (mr *MockBatchRecordReaderMockRecorder) Source() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Source", reflect.TypeOf((*MockBatchRecordReader)(nil).Source))
}
