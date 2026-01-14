// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package team

import (
	"context"
	"encoding/json"
	"fmt"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 验证邀请码信息
func NewVerifyInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyInviteLogic {
	return &VerifyInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyInviteLogic) VerifyInvite(req *types.VerifyInviteReq) (resp *types.VerifyInviteResp, err error) {
	// 1. 从 Redis 获取邀请码信息
	redisKey := fmt.Sprintf("rag:invite:code:%s", req.Code)
	data, err := l.svcCtx.RedisClient.GetCtx(l.ctx, redisKey)
	if err != nil || data == "" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "邀请码无效或已过期")
	}

	// 2. 解析邀请数据
	var inviteData InviteData
	if err := json.Unmarshal([]byte(data), &inviteData); err != nil {
		l.Errorf("解析邀请数据失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "邀请码数据异常")
	}

	// 3. 查询租户信息
	tenant, err := l.svcCtx.TenantModel.FindOne(l.ctx, inviteData.TenantId)
	if err != nil {
		l.Errorf("查询租户信息失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "租户信息不存在")
	}

	// 4. 查询邀请人信息
	inviter, err := l.svcCtx.UserModel.FindOne(l.ctx, inviteData.InviterId)
	if err != nil {
		l.Errorf("查询邀请人信息失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "邀请人信息不存在")
	}

	// 获取租户名称
	tenantName := ""
	if tenant.Name.Valid {
		tenantName = tenant.Name.String
	}

	return &types.VerifyInviteResp{
		TenantId:   inviteData.TenantId,
		TenantName: tenantName,
		Inviter:    inviter.Nickname,
	}, nil
}
