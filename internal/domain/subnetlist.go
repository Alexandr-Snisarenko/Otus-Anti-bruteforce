package domain

import (
	"context"
	"errors"
	"net"
)

// SubnetList представляет список подсетей в CIDR нотации.
type SubnetList struct {
	nets map[string]*net.IPNet
}

func NewSubnetList() *SubnetList {
	return &SubnetList{nets: make(map[string]*net.IPNet)}
}

func (nl *SubnetList) Add(cidr string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.Join(ErrInvalidCIDR, err)
	}

	// Проверка корректности CIDR. IP адрес должен быть адресом подсети.
	if !ip.Equal(ipnet.IP) {
		return ErrInvalidCIDR
	}

	nl.nets[cidr] = ipnet
	return nil
}

func (nl *SubnetList) Contains(ipStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, ErrInvalidIP
	}

	for _, net := range nl.nets {
		if net.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

func (nl *SubnetList) Remove(cidr string) {
	delete(nl.nets, cidr)
}

func (nl *SubnetList) Clear() {
	nl.nets = make(map[string]*net.IPNet)
}

func (nl *SubnetList) Load(ctx context.Context, repo SubnetRepo, listType ListType) ([]string, error) {
	cidrs, err := repo.GetSubnetLists(ctx, listType)
	if err != nil {
		return nil, err
	}

	nl.Clear()
	for _, cidr := range cidrs {
		if err := nl.Add(cidr); err != nil {
			return nil, err
		}
	}

	return cidrs, nil
}
