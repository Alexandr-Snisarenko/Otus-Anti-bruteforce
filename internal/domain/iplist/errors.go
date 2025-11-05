package iplist

import (
	"errors"
)

var (
	ErrInvalidIP   = errors.New("invalid IP address")
	ErrInvalidCIDR = errors.New("invalid CIDR notation")
)
