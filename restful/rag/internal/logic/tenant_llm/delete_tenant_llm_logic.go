// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant_llm

import (
	"context"

	"gozero-rag/internal/model/tenant_llm"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantLlmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除租户LLM配置
func NewDeleteTenantLlmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLlmLogic {
	return &DeleteTenantLlmLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantLlmLogic) DeleteTenantLlm(req *types.DeleteTenantLlmReq) error {
	// 1. 从 JWT 获取租户ID (多租户隔离)
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		l.Errorf("获取租户ID失败: %v", err)
		return xerr.NewErrCodeMsg(xerr.Unauthorized, "获取租户信息失败")
	}

	// 2. 执行多租户安全删除 (只能删除属于当前租户的记录)
	err = l.svcCtx.TenantLlmModel.DeleteByIdAndTenantId(l.ctx, uint64(req.Id), tenantId)
	if err != nil {
		if err == tenant_llm.ErrNotFound {
			return xerr.NewErrCodeMsg(xerr.UserApiNotFoundError, "配置不存在或无权删除")
		}
		l.Errorf("删除配置失败: %v", err)
		return xerr.NewErrCodeMsg(xerr.InternalError, "删除配置失败")
	}

	l.Infof("租户 %s 删除配置 ID=%d 成功", tenantId, req.Id)
	return nil
}
