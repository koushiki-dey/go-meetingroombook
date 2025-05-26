package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

func ValidateTimeFormat(t time.Time) error {
	if t.IsZero() {
		return errors.New("time cannot be zero value")
	}

	return nil
}

func IsTimeConflict(start1, end1, start2, end2 time.Time) (bool, error) {
	if start1.IsZero() || end1.IsZero() || start2.IsZero() || end2.IsZero() {
		return false, errors.New("one or more times are zero")
	}

	if start1.Before(end2) && start2.Before(end1) {
		return true, nil
	}
	return false, nil
}

func ParseBody(r *http.Request, x interface{}) {
	if body, err := io.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal([]byte(body), x); err != nil {
			return
		}
	}
}

func IsCapacityExceeding(currentCapacity, maxCapacity int) (bool, error) {
	var ErrCapacityExceeded = errors.New("capacity exceeded")
	if currentCapacity > maxCapacity {
		return true, ErrCapacityExceeded
	}
	return false, nil
}
