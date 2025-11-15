package app

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/config"
	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/domain"
	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/domain/ratelimit"
	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/domain/subnetlist"
	"github.com/Alexandr-Snisarenko/Otus-Anti-bruteforce/internal/ports"
)

type RateLimiterService struct {
	whitelistSubnet *subnetlist.SubnetList
	blacklistSubnet *subnetlist.SubnetList
	limitChecker    *ratelimit.LimitChecker
}

func NewRateLimiterService(ctx context.Context, config *config.Limits, limiterRepo ports.LimiterRepo, subnetRepo ports.SubnetRepo) (*RateLimiterService, error) {
	limits := ratelimit.Limits{
		domain.LoginLimit: {
			Limit:  config.LoginAttempts,
			Window: config.Window,
		},
		domain.PasswordLimit: {
			Limit:  config.PasswordAttempts,
			Window: config.Window,
		},
		domain.IPLimit: {
			Limit:  config.IPAttempts,
			Window: config.Window,
		},
	}

	whitelistSubnet := subnetlist.NewSubnetList(domain.Whitelist)
	if err := whitelistSubnet.Load(ctx, subnetRepo); err != nil {
		return nil, err
	}

	blacklistSubnet := subnetlist.NewSubnetList(domain.Blacklist)
	if err := blacklistSubnet.Load(ctx, subnetRepo); err != nil {
		return nil, err
	}

	return &RateLimiterService{
		whitelistSubnet: whitelistSubnet,
		blacklistSubnet: blacklistSubnet,
		limitChecker:    ratelimit.NewLimitChecker(limiterRepo, limits),
	}, nil
}

func (s *RateLimiterService) Check(ctx context.Context, login, password, ip string) (bool, error) {
	// Проверка whitelist/blacklist
	if ok, err := s.whitelistSubnet.Contains(ip); err != nil || ok {
		return ok, err
	}
	if ok, err := s.blacklistSubnet.Contains(ip); err != nil || ok {
		return ok, err
	}

	// Хешируем пароль перед проверкой лимитов
	passwordHash := hashPassword(password)

	// Три проверки: по логину, паролю, IP
	if okLogin, err := s.limitChecker.Allow(ctx, domain.LoginLimit, login); err != nil || !okLogin {
		return false, nil
	}
	if okPassword, err := s.limitChecker.Allow(ctx, domain.PasswordLimit, passwordHash); err != nil || !okPassword {
		return false, nil
	}
	if okIP, err := s.limitChecker.Allow(ctx, domain.IPLimit, ip); err != nil || !okIP {
		return false, nil
	}

	return true, nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}
