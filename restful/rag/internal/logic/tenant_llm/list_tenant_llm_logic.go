// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant_llm

import (
	"context"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantLlmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取租户LLM配置列表
func NewListTenantLlmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantLlmLogic {
	return &ListTenantLlmLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantLlmLogic) ListTenantLlm(req *types.ListTenantLlmReq) (resp *types.ListTenantLlmResp, err error) {
	// 1. 从 JWT 获取租户ID (多租户隔离)
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		l.Errorf("获取租户ID失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "获取租户信息失败")
	}

	// 2. 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 3. 查询列表 (只能查询当前租户的数据)
	list, total, err := l.svcCtx.TenantLlmModel.FindListByTenantId(
		l.ctx,
		tenantId,
		req.LlmFactory,
		req.ModelType,
		req.Status,
		page,
		pageSize,
	)
	if err != nil {
		l.Errorf("查询租户LLM配置列表失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "查询配置列表失败")
	}

	// 4. 转换为响应类型
	items := make([]types.TenantLlmInfo, 0, len(list))
	for _, item := range list {
		// API Key 脱敏处理
		maskedApiKey := maskApiKey(item.ApiKey.String)

		items = append(items, types.TenantLlmInfo{
			Id:          int64(item.Id),
			TenantId:    item.TenantId,
			LlmFactory:  item.LlmFactory,
			ModelType:   item.ModelType.String,
			LlmName:     item.LlmName,
			ApiKey:      maskedApiKey,
			ApiBase:     item.ApiBase.String,
			MaxTokens:   item.MaxTokens,
			UsedTokens:  item.UsedTokens,
			Status:      item.Status,
			CreatedTime: item.CreatedTime,
			UpdatedTime: item.UpdatedTime,
		})
	}

	return &types.ListTenantLlmResp{
		Total: total,
		List:  items,
	}, nil
}

// maskApiKey 对 API Key 进行脱敏处理，只显示前4位和后4位
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}
