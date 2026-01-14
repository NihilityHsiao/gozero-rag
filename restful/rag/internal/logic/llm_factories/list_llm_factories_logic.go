// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package llm_factories

import (
	"context"
	"strings"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLlmFactoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取系统支持的LLM厂商列表
func NewListLlmFactoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLlmFactoriesLogic {
	return &ListLlmFactoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLlmFactoriesLogic) ListLlmFactories(req *types.ListLlmFactoriesReq) (resp *types.ListLlmFactoriesResp, err error) {
	// 查询所有有效的厂商列表
	factories, err := l.svcCtx.LlmFactoriesModel.FindAllActive(l.ctx)
	if err != nil {
		l.Errorf("查询LLM厂商列表失败: %v", err)
		return nil, err
	}

	// 转换为响应类型
	list := make([]types.LlmFactoryInfo, 0, len(factories))
	for _, factory := range factories {
		// 将 tags 字符串拆分为数组
		var tagList []string
		if factory.Tags != "" {
			tagList = strings.Split(factory.Tags, ",")
		}

		// 处理 Logo (可能为 NULL)
		logo := ""
		if factory.Logo.Valid {
			logo = factory.Logo.String
		}

		list = append(list, types.LlmFactoryInfo{
			Name:    factory.Name,
			Logo:    logo,
			Tags:    factory.Tags,
			TagList: tagList,
			Rank:    factory.Rank,
			Status:  factory.Status,
		})
	}

	return &types.ListLlmFactoriesResp{
		List: list,
	}, nil
}
