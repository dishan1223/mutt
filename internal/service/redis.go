package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/dishan1223/mutt/internal/config"
)

const (
	refreshTokenPrefix = "refresh:"
	blacklistPrefix    = "blacklist:"
	refreshTokenTTL    = 7 * 24 * time.Hour
	blacklistTTL       = 15 * time.Minute
)

func StoreRefreshToken(token string, userID uint) error {
	ctx := context.Background()
	key := refreshTokenPrefix + hashToken(token)
	return config.RDB.Set(ctx, key, userID, refreshTokenTTL).Err()
}

func GetRefreshTokenUserID(token string) (uint, error) {
	ctx := context.Background()
	key := refreshTokenPrefix + hashToken(token)
	val, err := config.RDB.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func DeleteRefreshToken(token string) error {
	ctx := context.Background()
	key := refreshTokenPrefix + hashToken(token)
	return config.RDB.Del(ctx, key).Err()
}

func BlacklistAccessToken(tokenID string) error {
	ctx := context.Background()
	key := blacklistPrefix + tokenID
	return config.RDB.Set(ctx, key, "1", blacklistTTL).Err()
}

func IsAccessTokenBlacklisted(tokenID string) (bool, error) {
	ctx := context.Background()
	key := blacklistPrefix + tokenID
	exists, err := config.RDB.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
