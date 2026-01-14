package jwt

import (
	"errors"
	"gozero-rag/internal/xerr"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// CustomClaims 自定义claims，包含用户信息
type CustomClaims struct {
	Uid      string `json:"uid"`       // 用户ID (UUID v7)
	TenantId string `json:"tenant_id"` // 当前租户ID
	Nickname string `json:"nickname"`  // 用户昵称
	jwt.RegisteredClaims
}

// JwtConfig JWT配置
type JwtConfig struct {
	AccessSecret string
	AccessExpire int64 // 过期时间，单位秒
}

// GenerateToken 生成JWT token
func GenerateToken(cfg JwtConfig, uid string, tenantId string, nickname string) (accessToken string, expireAt int64, err error) {
	now := time.Now()
	expireAt = now.Add(time.Duration(cfg.AccessExpire) * time.Second).Unix()

	claims := CustomClaims{
		Uid:      uid,
		TenantId: tenantId,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.AccessExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "rag-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		return "", 0, err
	}

	return accessToken, expireAt, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string, accessSecret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "token格式错误")

			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "token已过期")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "token还没生效")

			}
		}
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "token无效")
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "token无效")

}

// GenerateRefreshToken 生成刷新token（过期时间更长）
func GenerateRefreshToken(cfg JwtConfig, uid string, tenantId string, nickname string) (refreshToken string, err error) {
	// 刷新token过期时间是access token的7倍（例如7天）
	refreshExpire := cfg.AccessExpire * 7
	now := time.Now()

	claims := CustomClaims{
		Uid:      uid,
		TenantId: tenantId,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(refreshExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "rag-service-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err = token.SignedString([]byte(cfg.AccessSecret))
	return
}
