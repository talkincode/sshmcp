package sshclient

import (
	"errors"
	"io"
	"testing"
)

// mockCloser implements io.Closer for testing
type mockCloser struct {
	closeErr error
	closed   bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.closeErr
}

func TestCloseIgnore_NilCloser(t *testing.T) {
	var err error
	CloseIgnore(&err, nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCloseIgnore_NoError(t *testing.T) {
	var err error
	closer := &mockCloser{}
	CloseIgnore(&err, closer)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !closer.closed {
		t.Error("Expected closer to be closed")
	}
}

func TestCloseIgnore_ErrorNotIgnored(t *testing.T) {
	var err error
	testErr := errors.New("test error")
	closer := &mockCloser{closeErr: testErr}

	CloseIgnore(&err, closer)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, testErr) {
		t.Errorf("Expected error to be %v, got: %v", testErr, err)
	}
}

func TestCloseIgnore_ErrorIgnored(t *testing.T) {
	var err error
	testErr := errors.New("test error")
	closer := &mockCloser{closeErr: testErr}

	CloseIgnore(&err, closer, testErr)

	if err != nil {
		t.Errorf("Expected error to be ignored, got: %v", err)
	}
}

func TestCloseIgnore_MergeErrors(t *testing.T) {
	existingErr := errors.New("existing error")
	closeErr := errors.New("close error")

	err := existingErr
	closer := &mockCloser{closeErr: closeErr}

	CloseIgnore(&err, closer)

	if err == nil {
		t.Error("Expected merged error, got nil")
	}
	if !errors.Is(err, existingErr) {
		t.Error("Expected error to contain existing error")
	}
	if !errors.Is(err, closeErr) {
		t.Error("Expected error to contain close error")
	}
}

func TestCloseIgnore_NilErrPointer(t *testing.T) {
	testErr := errors.New("test error")
	closer := &mockCloser{closeErr: testErr}

	// Should not panic with nil error pointer
	CloseIgnore(nil, closer)

	if !closer.closed {
		t.Error("Expected closer to be closed")
	}
}

func TestCloseIgnore_WrappedError(t *testing.T) {
	var err error
	baseErr := io.ErrUnexpectedEOF
	wrappedErr := errors.New("wrapped: " + baseErr.Error())
	closer := &mockCloser{closeErr: wrappedErr}

	// Wrapped error should not match
	CloseIgnore(&err, closer, baseErr)

	if err == nil {
		t.Error("Expected error since wrapped error doesn't match")
	}
}
