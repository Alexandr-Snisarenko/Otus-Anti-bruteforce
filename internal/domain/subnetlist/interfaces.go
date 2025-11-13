package subnetlist

import "context"

type ListType string

const (
	Whitelist ListType = "whitelist"
	Blacklist ListType = "blacklist"
)

// SubnetRepo — абстракция для работы со списками подсетей.
type SubnetRepo interface {
	GetSubnetLists(ctx context.Context, listType ListType) ([]string, error)
	SaveSubnetList(ctx context.Context, listType ListType, cidrs []string) error
	ClearSubnetList(ctx context.Context, listType ListType) error
	AddCIDRToSubnetList(ctx context.Context, listType ListType, cidr string) error
	RemoveCIDRFromSubnetList(ctx context.Context, listType ListType, cidr string) error
}
