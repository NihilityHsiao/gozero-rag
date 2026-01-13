package common

import (
	"context"
	"encoding/json"
	"gozero-rag/internal/xerr"
)

// GetUidFromCtx  从 context 中获取用户ID
func GetUidFromCtx(ctx context.Context) (int64, error) {
	uid, ok := ctx.Value("uid").(json.Number)
	if !ok {
		return 0, xerr.NewErrCodeMsg(xerr.InternalError, "获取用户信息失败")
	}
	uidInt, err := uid.Int64()
	if err != nil {
		return 0, xerr.NewErrCodeMsg(xerr.InternalError, "用户ID格式错误")
	}
	return uidInt, nil
}

// GetUsernameFromCtx  从 context 中获取用户名
func GetUsernameFromCtx(ctx context.Context) (string, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return "", xerr.NewErrCodeMsg(xerr.InternalError, "获取用户名失败")
	}
	return username, nil
}
