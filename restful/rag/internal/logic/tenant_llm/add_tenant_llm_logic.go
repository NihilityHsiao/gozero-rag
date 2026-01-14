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

type AddTenantLlmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量添加租户LLM配置
func NewAddTenantLlmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddTenantLlmLogic {
	return &AddTenantLlmLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddTenantLlmLogic) AddTenantLlm(req *types.AddTenantLlmReq) (resp *types.AddTenantLlmResp, err error) {
	// 1. 从 JWT 获取租户ID (多租户隔离)
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		l.Errorf("获取租户ID失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "获取租户信息失败")
	}

	// 2. 验证请求参数
	if req.LlmFactory == "" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "厂商名称不能为空")
	}
	if req.ApiKey == "" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "API Key 不能为空")
	}
	if req.ApiBase == "" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "API Base 不能为空")
	}
	if len(req.Models) == 0 {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "请至少添加一个模型")
	}

	// 3. 验证厂商是否存在
	factory, err := l.svcCtx.LlmFactoriesModel.FindOne(l.ctx, req.LlmFactory)
	if err != nil {
		l.Errorf("查询厂商 %s 失败: %v", req.LlmFactory, err)
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "不支持的厂商: "+req.LlmFactory)
	}
	if factory.Status != 1 {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "该厂商已被禁用")
	}

	// 4. 构建待插入的记录
	records := make([]*tenant_llm.TenantLlm, 0, len(req.Models))
	for _, model := range req.Models {
		if model.LlmName == "" {
			continue // 跳过空的模型名称
		}

		maxTokens := model.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 8192 // 默认值
		}

		records = append(records, &tenant_llm.TenantLlm{
			TenantId:   tenantId,
			LlmFactory: req.LlmFactory,
			ModelType:  tenant_llm.ToNullString(model.ModelType),
			LlmName:    model.LlmName,
			ApiKey:     tenant_llm.ToNullString(req.ApiKey),
			ApiBase:    tenant_llm.ToNullString(req.ApiBase),
			MaxTokens:  maxTokens,
			UsedTokens: 0,
			Status:     1, // 默认启用
		})
	}

	if len(records) == 0 {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "没有有效的模型配置")
	}

	// 5. 批量插入 (忽略冲突)
	successCount, failedModels, err := l.svcCtx.TenantLlmModel.BatchInsertIgnore(l.ctx, records)
	if err != nil {
		l.Errorf("批量插入失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "添加模型配置失败")
	}

	l.Infof("租户 %s 添加 %s 厂商模型成功: %d 个, 失败: %d 个", tenantId, req.LlmFactory, successCount, len(failedModels))

	return &types.AddTenantLlmResp{
		SuccessCount: successCount,
		FailedCount:  int64(len(failedModels)),
		FailedModels: failedModels,
	}, nil
}
