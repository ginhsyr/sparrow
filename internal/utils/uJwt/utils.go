// Package uJwt 提供与 JWT（JSON Web Token）相关的通用工具方法，
// 包括生成和解析基于 HMAC 的 JWT
package uJwt

// @Title        utils.go
// @Description  JWT相关通用工具包

import (
	"Sparrow/configs"
	"Sparrow/internal/model"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Claims 定义了JWT字符串解析内容
type Claims struct {
	UserID string         `json:"userID"` // 用户ID
	Role   model.RoleType `json:"role"`   // 用户角色
	jwt.RegisteredClaims
}

// JWTParse 将JWT字符串解析至 *Claims 结构体
func JWTParse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return configs.JWTSigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// GenerateJWT 创建并签名一个新的 JWT 字符串
// 参数 userID 为用户 ID，role 为用户角色
// 返回生成的 JWT 字符串及可能出现的错误
// 生成的 token 有效期为 24 小时
func GenerateJWT(userID string, role model.RoleType) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sparrow",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(configs.JWTSigningKey)
}
