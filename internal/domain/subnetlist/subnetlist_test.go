package subnetlist

import (
	"errors"
	"testing"

	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/domain"
)

func TestIsIPInList_Table(t *testing.T) {
	tests := []struct {
		name  string
		CIDRs []string
		ip    string
		want  bool
		err   error
	}{
		{
			name:  "ipv4 found",
			CIDRs: []string{"192.168.0.0/24"},

			ip:   "192.168.0.10",
			want: true,
			err:  nil,
		},
		{
			name:  "ipv4 not found",
			CIDRs: []string{"10.0.0.0/8"},
			ip:    "192.168.1.1",
			want:  false,
			err:   nil,
		},
		{
			name:  "empty list",
			CIDRs: []string{},
			ip:    "1.2.3.4",
			want:  false,
			err:   nil,
		},
		{
			name:  "invalid ip",
			CIDRs: []string{"127.0.0.0/8"},
			ip:    "not-an-ip",
			want:  false,
			err:   ErrInvalidIP,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list := NewSubnetList(domain.Whitelist)
			for _, cidr := range tc.CIDRs {
				err := list.Add(cidr)
				if err != nil && !errors.Is(err, tc.err) {
					t.Fatalf("unexpected error adding cidr %s: %v", cidr, err)
				}
			}
			got, err := list.Contains(tc.ip)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected err %v, got %v", tc.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestAdd_InvalidCIDR(t *testing.T) {
	tests := []struct {
		name string
		CIDR string
	}{
		{
			name: "not-an-CIDR",
			CIDR: "invalid-cidr",
		},
		{
			name: "missing-slash",
			CIDR: "192.168.0.0",
		},
		{
			name: "not-valid-CIDR",
			CIDR: "10.10.10.10/24",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list := NewSubnetList(domain.Whitelist)
			err := list.Add(tc.CIDR)
			if err == nil {
				t.Fatalf("expected error for invalid CIDR, got nil")
			}
			if !errors.Is(err, ErrInvalidCIDR) {
				t.Fatalf("expected ErrInvalidCIDR, got %v", err)
			}
		})
	}
}
