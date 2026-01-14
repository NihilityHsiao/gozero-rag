package common

import (
	"context"
	"gozero-rag/internal/xerr"
)

// GetUidFromCtx 从 context 中获取用户ID (UUID字符串)
func GetUidFromCtx(ctx context.Context) (string, error) {
	uid, ok := ctx.Value("uid").(string)
	if !ok {
		return "", xerr.NewErrCodeMsg(xerr.InternalError, "获取用户信息失败")
	}
	return uid, nil
}

// GetTenantIdFromCtx 从 context 中获取当前租户ID
func GetTenantIdFromCtx(ctx context.Context) (string, error) {
	tenantId, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return "", xerr.NewErrCodeMsg(xerr.InternalError, "获取租户信息失败")
	}
	return tenantId, nil
}

// GetNicknameFromCtx 从 context 中获取用户昵称
func GetNicknameFromCtx(ctx context.Context) (string, error) {
	nickname, ok := ctx.Value("nickname").(string)
	if !ok {
		return "", xerr.NewErrCodeMsg(xerr.InternalError, "获取用户昵称失败")
	}
	return nickname, nil
}
