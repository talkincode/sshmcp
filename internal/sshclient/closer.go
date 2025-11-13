package sshclient

import (
	"errors"
	"io"
)

// CloseIgnore closes the closer and handles errors appropriately.
// It will ignore errors in the ignore list and merge other errors into errp.
// This is useful for defer statements where we want to handle close errors
// but allow certain expected errors to be silently ignored.
//
// Usage example:
//
//	func handle(r io.ReadCloser) (err error) {
//	    defer CloseIgnore(&err, r, net.ErrClosed)
//	    // ... read ...
//	    return
//	}
func CloseIgnore(errp *error, c io.Closer, ignore ...error) {
	if c == nil {
		return
	}

	if cerr := c.Close(); cerr != nil {
		// Check if the error is in the ignore list
		for _, ig := range ignore {
			if errors.Is(cerr, ig) {
				return
			}
		}

		// Not in ignore list: merge into return error
		if errp != nil {
			if *errp == nil {
				*errp = cerr
			} else {
				*errp = errors.Join(*errp, cerr)
			}
		}
	}
}
