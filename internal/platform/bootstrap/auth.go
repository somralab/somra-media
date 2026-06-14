package bootstrap

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"

	"github.com/somralab/somra-media/internal/auth"
	"github.com/somralab/somra-media/internal/platform/config"
	"github.com/somralab/somra-media/internal/platform/db"
)

// AuthBundle groups auth dependencies wired for the API gateway.
type AuthBundle struct {
	Service *auth.Service
	Users   *db.UserRepo
}

// WireAuth builds auth services from platform components and config.
func WireAuth(c *Components, cfg config.AuthConfig, logger *slog.Logger) (*AuthBundle, error) {
	if c == nil || c.DB == nil {
		return nil, fmt.Errorf("bootstrap auth: db required")
	}
	secret := []byte(cfg.JWTSecret)
	if len(secret) < 32 {
		if logger != nil {
			logger.Warn("auth: SOMRA_JWT_SECRET missing or short; generating ephemeral secret (tokens invalid after restart)")
		}
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("bootstrap auth secret: %w", err)
		}
		secret = b
	}
	pepper := []byte(cfg.RefreshPepper)
	if len(pepper) < 16 {
		pepper = secret
	}

	q := c.DB.Querier()
	users := db.NewUserRepo(q)
	sessions := db.NewSessionRepo(q)
	profiles := db.NewProfileRepo(q)
	attempts := db.NewLoginAttemptRepo(q)

	jwtCfg := auth.DefaultJWTConfig(secret)
	if cfg.AccessTTL > 0 {
		jwtCfg.AccessTTL = cfg.AccessTTL
	}
	if cfg.RefreshTTL > 0 {
		jwtCfg.RefreshTTL = cfg.RefreshTTL
	}

	tokens := auth.NewJWTService(jwtCfg)
	refresh := auth.NewSQLiteRefreshStore(sessions, pepper, jwtCfg.RefreshTTL)
	hasher := auth.NewPasswordHasher(auth.DefaultPasswordPolicy())
	lockout := auth.NewLoginLockout(attempts, auth.DefaultLockoutConfig())

	svc := auth.NewService(auth.ServiceConfig{
		Users:    users,
		Sessions: sessions,
		Profiles: profiles,
		Tokens:   tokens,
		Refresh:  refresh,
		Hasher:   hasher,
		Lockout:  lockout,
		JWT:      jwtCfg,
	})

	return &AuthBundle{Service: svc, Users: users}, nil
}

// DevJWTSecret reads SOMRA_JWT_SECRET for local dev when set.
func DevJWTSecret() string {
	return os.Getenv("SOMRA_JWT_SECRET")
}
