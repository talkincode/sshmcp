package errutil

import (
	"errors"
	"io"
	"testing"
)

type mockCloser struct {
	closeErr error
	closed   bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.closeErr
}

func TestIsIgnorableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, true},
		{"EOF error", io.EOF, true},
		{"EOF string", errors.New("EOF"), true},
		{"normal error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIgnorableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsIgnorableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsEOFError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"EOF error", io.EOF, true},
		{"EOF string", errors.New("EOF"), true},
		{"normal error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEOFError(tt.err)
			if result != tt.expected {
				t.Errorf("IsEOFError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestSafeClose(t *testing.T) {
	tests := []struct {
		name      string
		closer    io.Closer
		expectErr bool
	}{
		{"nil closer", nil, false},
		{"successful close", &mockCloser{}, false},
		{"ignorable error", &mockCloser{closeErr: io.EOF}, false},
		{"non-ignorable error", &mockCloser{closeErr: errors.New("close failed")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeClose(tt.closer)
			if (err != nil) != tt.expectErr {
				t.Errorf("SafeClose() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestHandleCloseError(t *testing.T) {
	t.Run("nil closer", func(t *testing.T) {
		var err error
		HandleCloseError(&err, nil)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("successful close", func(t *testing.T) {
		var err error
		closer := &mockCloser{}
		HandleCloseError(&err, closer)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
		if !closer.closed {
			t.Error("Closer was not closed")
		}
	})

	t.Run("ignorable error", func(t *testing.T) {
		var err error
		closer := &mockCloser{closeErr: io.EOF}
		HandleCloseError(&err, closer)
		if err != nil {
			t.Errorf("Expected nil error for ignorable EOF, got %v", err)
		}
	})

	t.Run("merge with existing error", func(t *testing.T) {
		existingErr := errors.New("existing error")
		closeErr := errors.New("close error")
		err := existingErr
		closer := &mockCloser{closeErr: closeErr}
		HandleCloseError(&err, closer)
		if err == nil {
			t.Error("Expected merged error, got nil")
		}
		if !errors.Is(err, existingErr) {
			t.Error("Merged error should contain existing error")
		}
		if !errors.Is(err, closeErr) {
			t.Error("Merged error should contain close error")
		}
	})
}

func TestJoinErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	tests := []struct {
		name     string
		errors   []error
		expected int // expected number of errors
	}{
		{"no errors", []error{}, 0},
		{"one error", []error{err1}, 1},
		{"two errors", []error{err1, err2}, 2},
		{"with nil", []error{err1, nil, err2}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinErrors(tt.errors...)
			if tt.expected == 0 {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected error, got nil")
				}
			}
		})
	}
}

func TestEnhanceError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		result := EnhanceError(nil, "stdout", "stderr")
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("EOF with output", func(t *testing.T) {
		result := EnhanceError(io.EOF, "some output", "")
		if result != nil {
			t.Errorf("Expected nil for EOF with output, got %v", result)
		}
	})

	t.Run("EOF without output", func(t *testing.T) {
		result := EnhanceError(io.EOF, "", "")
		if result == nil {
			t.Error("Expected error for EOF without output")
		}
	})

	t.Run("regular error", func(t *testing.T) {
		testErr := errors.New("test error")
		result := EnhanceError(testErr, "stdout", "stderr")
		if result == nil {
			t.Error("Expected enhanced error")
		}
	})
}
