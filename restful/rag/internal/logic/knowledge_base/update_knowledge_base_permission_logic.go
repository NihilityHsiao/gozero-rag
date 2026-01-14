// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_base

import (
	"context"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateKnowledgeBasePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新知识库权限
func NewUpdateKnowledgeBasePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateKnowledgeBasePermissionLogic {
	return &UpdateKnowledgeBasePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateKnowledgeBasePermissionLogic) UpdateKnowledgeBasePermission(req *types.UpdateKnowledgeBasePermissionReq) (resp *types.UpdateKnowledgeBasePermissionResp, err error) {
	// 1. 获取当前用户和租户
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 查询知识库
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.Id)
	if err != nil {
		l.Errorf("查询知识库失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
	}

	// 3. 租户隔离校验
	if kb.TenantId != tenantId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权限访问该知识库")
	}

	// 4. Owner 权限校验：只有创建者可以修改权限
	if kb.CreatedBy != userId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "只有知识库Owner可以修改权限")
	}

	// 5. 权限值校验
	if req.Permission != "me" && req.Permission != "team" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "权限值只能为 me 或 team")
	}

	// 6. 更新权限（只更新 permission 字段）
	err = l.svcCtx.KnowledgeBaseModel.UpdatePermission(l.ctx, req.Id, req.Permission)
	if err != nil {
		l.Errorf("更新知识库权限失败: %v", err)
		return nil, xerr.NewInternalErrMsg("更新权限失败")
	}

	return &types.UpdateKnowledgeBasePermissionResp{}, nil
}
