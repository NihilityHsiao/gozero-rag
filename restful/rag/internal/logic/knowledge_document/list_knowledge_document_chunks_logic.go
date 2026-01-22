// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKnowledgeDocumentChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文档切片列表
func NewListKnowledgeDocumentChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKnowledgeDocumentChunksLogic {
	return &ListKnowledgeDocumentChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListKnowledgeDocumentChunksLogic) ListKnowledgeDocumentChunks(req *types.ListKnowledgeDocumentChunksReq) (resp *types.ListKnowledgeDocumentChunksResp, err error) {
	// 从 ES 查询切片列表
	result, err := l.svcCtx.ChunkModel.ListByDocId(
		l.ctx,
		req.Id, // 文档ID
		req.Keyword,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		l.Errorf("查询切片列表失败: %v", err)
		return nil, err
	}

	// 组装响应
	list := make([]types.ChunkInfo, 0, len(result.Chunks))
	for _, chunk := range result.Chunks {
		list = append(list, types.ChunkInfo{
			Id:          chunk.Id,
			Content:     chunk.Content,
			DocId:       chunk.DocId,
			DocName:     chunk.DocName,
			ImportantKw: chunk.ImportantKw,
			CreatedAt:   chunk.CreateTime,
		})
	}

	resp = &types.ListKnowledgeDocumentChunksResp{
		Total: result.Total,
		List:  list,
	}

	return resp, nil
}
