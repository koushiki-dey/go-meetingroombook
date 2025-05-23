package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

func ValidateTimeFormat(timeStr string) error {
	_, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return errors.New("time must be in RFC3339 format i.e. yyyy-MM-ddThh:mm:ss.sssZ (e.g., 2025-05-22T15:00:00Z)")
	}
	return nil
}

func IsTimeConflict(start1, end1, start2, end2 string) (bool, error) {
	t1Start, err := time.Parse(time.RFC3339, start1)
	if err != nil {
		return false, err
	}
	t1End, err := time.Parse(time.RFC3339, end1)
	if err != nil {
		return false, err
	}
	t2Start, err := time.Parse(time.RFC3339, start2)
	if err != nil {
		return false, err
	}
	t2End, err := time.Parse(time.RFC3339, end2)
	if err != nil {
		return false, err
	}

	if t1Start.Before(t2End) && t2Start.Before(t1End) {
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
