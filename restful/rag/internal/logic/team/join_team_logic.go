// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package team

import (
	"context"
	"encoding/json"
	"fmt"

	"gozero-rag/internal/model/user_tenant"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type JoinTeamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 使用邀请码加入团队
func NewJoinTeamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinTeamLogic {
	return &JoinTeamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JoinTeamLogic) JoinTeam(req *types.JoinTeamReq) (resp *types.JoinTeamResp, err error) {
	// 1. 从 context 获取当前用户ID
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 从 Redis 获取邀请码信息
	redisKey := fmt.Sprintf("rag:invite:code:%s", req.InviteCode)
	data, err := l.svcCtx.RedisClient.GetCtx(l.ctx, redisKey)
	if err != nil || data == "" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "邀请码无效或已过期")
	}

	// 3. 解析邀请数据
	var inviteData InviteData
	if err := json.Unmarshal([]byte(data), &inviteData); err != nil {
		l.Errorf("解析邀请数据失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "邀请码数据异常")
	}

	// 4. 检查用户是否已加入该租户
	_, err = l.svcCtx.UserTenantModel.FindByUserIdAndTenantId(l.ctx, userId, inviteData.TenantId)
	if err == nil {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "您已是该团队成员")
	}

	// 5. 创建 user_tenant 关联记录
	newId, _ := uuid.NewV7()
	userTenantRecord := &user_tenant.UserTenant{
		Id:        newId.String(),
		UserId:    userId,
		TenantId:  inviteData.TenantId,
		Role:      inviteData.Role,
		InvitedBy: inviteData.InviterId,
		Status:    1,
	}

	_, err = l.svcCtx.UserTenantModel.Insert(l.ctx, userTenantRecord)
	if err != nil {
		l.Errorf("创建用户租户关联失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "加入团队失败")
	}

	// 6. 删除 Redis 邀请码 (一次性邀请码)
	_, err = l.svcCtx.RedisClient.DelCtx(l.ctx, redisKey)
	if err != nil {
		l.Errorf("删除邀请码失败: %v", err)
		// 不影响主流程，只记录日志
	}

	return &types.JoinTeamResp{
		Success: true,
	}, nil
}
