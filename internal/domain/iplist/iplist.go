package iplist

import (
	"net"
)

type IPlist struct {
	CIDRs []string
}

func (l *IPlist) IsIPInList(ipStr string) (bool, error) {
	for _, cidr := range l.CIDRs {
		if ok, err := isIPInCIDR(ipStr, cidr); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func isIPInCIDR(ipStr, cidr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, ErrInvalidIP
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, ErrInvalidCIDR
	}

	return network.Contains(ip), nil
}
