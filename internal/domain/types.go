package domain

// ListType представляет тип списка подсетей.
type ListType string

const (
	Whitelist ListType = "whitelist"
	Blacklist ListType = "blacklist"
)

// LimitType представляет тип ограничения.
type LimitType string

const (
	LoginLimit    LimitType = "login"
	PasswordLimit LimitType = "password"
	IPLimit       LimitType = "ip"
)
