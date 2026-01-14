// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant_llm

import (
	"context"
	"strings"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantLlmGroupedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取租户LLM配置列表(按厂商分组)
func NewListTenantLlmGroupedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantLlmGroupedLogic {
	return &ListTenantLlmGroupedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantLlmGroupedLogic) ListTenantLlmGrouped(req *types.ListTenantLlmReq) (resp *types.ListTenantLlmGroupedResp, err error) {
	// 1. 从 JWT 获取租户ID (多租户隔离)
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		l.Errorf("获取租户ID失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "获取租户信息失败")
	}

	// 2. 按厂商分组查询 (只能查询当前租户的数据)
	groupedData, err := l.svcCtx.TenantLlmModel.FindGroupedByFactory(l.ctx, tenantId)
	if err != nil {
		l.Errorf("查询按厂商分组的配置列表失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "查询配置列表失败")
	}

	// 3. 获取所有厂商信息 (用于获取 Logo)
	factories, err := l.svcCtx.LlmFactoriesModel.FindAllActive(l.ctx)
	if err != nil {
		l.Errorf("查询厂商列表失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "查询厂商信息失败")
	}

	// 构建厂商名称到 Logo 的映射
	factoryLogoMap := make(map[string]string)
	for _, f := range factories {
		if f.Logo.Valid {
			factoryLogoMap[f.Name] = f.Logo.String
		}
	}

	// 4. 转换为响应类型
	list := make([]types.TenantLlmGroupByFactory, 0, len(groupedData))
	for factoryName, items := range groupedData {
		models := make([]types.TenantLlmInfo, 0, len(items))
		var apiBase string

		for _, item := range items {
			// API Key 脱敏处理
			maskedApiKey := maskApiKey(item.ApiKey.String)

			// 记录第一个模型的 api_base (同一厂商的模型共享同一 api_base)
			if apiBase == "" && item.ApiBase.Valid {
				apiBase = item.ApiBase.String
			}

			models = append(models, types.TenantLlmInfo{
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

		list = append(list, types.TenantLlmGroupByFactory{
			LlmFactory:  factoryName,
			FactoryLogo: factoryLogoMap[factoryName],
			ApiBase:     maskApiBase(apiBase),
			Models:      models,
		})
	}

	return &types.ListTenantLlmGroupedResp{
		List: list,
	}, nil
}

// maskApiBase 对 API Base 进行简化显示
func maskApiBase(apiBase string) string {
	if apiBase == "" {
		return ""
	}
	// 只显示域名部分
	if strings.HasPrefix(apiBase, "https://") {
		parts := strings.Split(strings.TrimPrefix(apiBase, "https://"), "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	if strings.HasPrefix(apiBase, "http://") {
		parts := strings.Split(strings.TrimPrefix(apiBase, "http://"), "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return apiBase
}
