// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package team

import (
	"context"

	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListJoinedTeamsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取我加入的团队列表
func NewListJoinedTeamsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListJoinedTeamsLogic {
	return &ListJoinedTeamsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListJoinedTeamsLogic) ListJoinedTeams() (resp *types.ListJoinedTeamsResp, err error) {
	// 从 context 获取当前用户ID
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询用户加入的所有租户
	userTenants, err := l.svcCtx.UserTenantModel.FindByUserId(l.ctx, userId)
	if err != nil {
		l.Errorf("查询加入的团队失败: %v", err)
		return nil, err
	}

	// 转换为响应结构
	list := make([]types.JoinedTeam, 0, len(userTenants))
	for _, ut := range userTenants {
		// 获取每个租户的 Owner 信息
		owner, err := l.svcCtx.UserTenantModel.FindOwnerByTenantId(l.ctx, ut.TenantId)
		ownerName := ""
		if err == nil && owner != nil {
			ownerName = owner.Nickname
		}

		list = append(list, types.JoinedTeam{
			TenantId:   ut.TenantId,
			TenantName: ut.TenantName,
			OwnerName:  ownerName,
			Role:       ut.Role,
			JoinedTime: ut.CreatedTime,
		})
	}

	return &types.ListJoinedTeamsResp{
		List: list,
	}, nil
}
