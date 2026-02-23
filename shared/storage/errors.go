package storage

import (
	"errors"

	"github.com/lib/pq"
)

type pgCode string

const (
	UniqueViolationCode pgCode = "23505"
)

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
