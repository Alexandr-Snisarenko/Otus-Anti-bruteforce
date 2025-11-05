package iplist

import (
	"errors"
	"testing"
)

func TestIsIPInList_Table(t *testing.T) {
	tests := []struct {
		name string
		list IPlist
		ip   string
		want bool
		err  error
	}{
		{
			name: "happy path - ipv4 found",
			list: IPlist{CIDRs: []string{"192.168.0.0/24"}},
			ip:   "192.168.0.10",
			want: true,
			err:  nil,
		},
		{
			name: "ipv4 not found",
			list: IPlist{CIDRs: []string{"10.0.0.0/8"}},
			ip:   "192.168.1.1",
			want: false,
			err:  nil,
		},
		{
			name: "empty list",
			list: IPlist{CIDRs: []string{}},
			ip:   "1.2.3.4",
			want: false,
			err:  nil,
		},
		{
			name: "invalid ip",
			list: IPlist{CIDRs: []string{"127.0.0.0/8"}},
			ip:   "not-an-ip",
			want: false,
			err:  ErrInvalidIP,
		},
		{
			name: "invalid cidr",
			list: IPlist{CIDRs: []string{"not-a-cidr"}},
			ip:   "127.0.0.1",
			want: false,
			err:  ErrInvalidCIDR,
		},
		{
			name: "order - first invalid then valid",
			list: IPlist{CIDRs: []string{"not-a-cidr", "192.168.0.0/24"}},
			ip:   "192.168.0.1",
			want: false,
			err:  ErrInvalidCIDR,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.list.IsIPInList(tc.ip)
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

func Test_isIPInCIDR_IPv6(t *testing.T) {
	ok, err := isIPInCIDR("::1", "::1/128")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected IPv6 loopback to be contained in ::1/128")
	}
}
