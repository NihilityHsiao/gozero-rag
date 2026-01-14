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

type ListMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前团队成员列表
func NewListMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMembersLogic {
	return &ListMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMembersLogic) ListMembers() (resp *types.ListMembersResp, err error) {
	// 从 context 获取当前租户ID
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询租户下所有成员
	members, err := l.svcCtx.UserTenantModel.FindMembersByTenantId(l.ctx, tenantId)
	if err != nil {
		l.Errorf("查询团队成员失败: %v", err)
		return nil, err
	}

	// 转换为响应结构
	list := make([]types.TeamMember, 0, len(members))
	for _, m := range members {
		list = append(list, types.TeamMember{
			UserId:     m.UserId,
			Nickname:   m.Nickname,
			Email:      m.Email,
			Role:       m.Role,
			Status:     m.Status,
			JoinedTime: m.CreatedTime,
		})
	}

	return &types.ListMembersResp{
		List: list,
	}, nil
}
