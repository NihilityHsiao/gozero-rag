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

type UpdateTenantLlmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新租户LLM配置
func NewUpdateTenantLlmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLlmLogic {
	return &UpdateTenantLlmLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantLlmLogic) UpdateTenantLlm(req *types.UpdateTenantLlmReq) error {
	// 1. 从 JWT 获取租户ID (多租户隔离)
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		l.Errorf("获取租户ID失败: %v", err)
		return xerr.NewErrCodeMsg(xerr.Unauthorized, "获取租户信息失败")
	}

	// 2. 验证记录是否存在且属于当前租户 (多租户安全)
	existing, err := l.svcCtx.TenantLlmModel.FindOneByIdAndTenantId(l.ctx, uint64(req.Id), tenantId)
	if err != nil {
		if err == tenant_llm.ErrNotFound {
			return xerr.NewErrCodeMsg(xerr.UserApiNotFoundError, "配置不存在或无权访问")
		}
		l.Errorf("查询配置失败: %v", err)
		return xerr.NewErrCodeMsg(xerr.InternalError, "查询配置失败")
	}

	// 3. 构建更新数据 (只更新允许修改的字段)
	updateData := &tenant_llm.TenantLlm{
		Id:         existing.Id,
		TenantId:   tenantId,            // 保持租户ID不变
		LlmFactory: existing.LlmFactory, // 厂商不可变
		ModelType:  existing.ModelType,  // 类型不可变
		LlmName:    existing.LlmName,    // 模型名称不可变
		UsedTokens: existing.UsedTokens, // 保持已使用Token数
	}

	// 只更新传入的字段
	if req.ApiKey != "" {
		updateData.ApiKey = tenant_llm.ToNullString(req.ApiKey)
	} else {
		updateData.ApiKey = existing.ApiKey
	}

	if req.ApiBase != "" {
		updateData.ApiBase = tenant_llm.ToNullString(req.ApiBase)
	} else {
		updateData.ApiBase = existing.ApiBase
	}

	if req.MaxTokens > 0 {
		updateData.MaxTokens = req.MaxTokens
	} else {
		updateData.MaxTokens = existing.MaxTokens
	}

	if req.Status > 0 {
		updateData.Status = req.Status
	} else {
		updateData.Status = existing.Status
	}

	// 4. 执行更新 (使用多租户安全的更新方法)
	err = l.svcCtx.TenantLlmModel.UpdateByIdAndTenantId(l.ctx, updateData)
	if err != nil {
		l.Errorf("更新配置失败: %v", err)
		return xerr.NewErrCodeMsg(xerr.InternalError, "更新配置失败")
	}

	l.Infof("租户 %s 更新配置 ID=%d 成功", tenantId, req.Id)
	return nil
}
