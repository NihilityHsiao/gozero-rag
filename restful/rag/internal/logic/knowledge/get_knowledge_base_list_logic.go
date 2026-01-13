// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeBaseListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库列表
func NewGetKnowledgeBaseListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeBaseListLogic {
	return &GetKnowledgeBaseListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeBaseListLogic) GetKnowledgeBaseList(req *types.GetKnowledgeBaseListReq) (resp *types.GetKnowledgeBaseListResp, err error) {
	count, err := l.svcCtx.KnowledgeBaseModel.Count(l.ctx, req.Status)
	if err != nil {
		l.Logger.Errorf("GetKnowledgeBaseList count error: %v", err)
		return nil, err
	}

	list, err := l.svcCtx.KnowledgeBaseModel.FindList(l.ctx, req.Page, req.PageSize, req.Status)
	if err != nil {
		l.Logger.Errorf("GetKnowledgeBaseList find list error: %v", err)
		return nil, err
	}

	respList := make([]types.KnowledgeBaseInfo, 0, len(list))
	for _, item := range list {
		respList = append(respList, types.KnowledgeBaseInfo{
			Id:          int64(item.Id),
			Name:        item.Name,
			Description: item.Description.String,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetKnowledgeBaseListResp{
		Total: count,
		List:  respList,
	}, nil
}
