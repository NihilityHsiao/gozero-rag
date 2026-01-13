// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"encoding/json"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeDocumentChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定知识库、指定文档的所有分片
func NewGetKnowledgeDocumentChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeDocumentChunksLogic {
	return &GetKnowledgeDocumentChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeDocumentChunksLogic) GetKnowledgeDocumentChunks(req *types.GetKnowledgeDocumentChunksReq) (resp *types.GetKnowledgeDocumentChunksResp, err error) {
	// Call model to get chunks
	chunks, count, err := l.svcCtx.KnowledgeDocumentChunkModel.FindListByDocId(l.ctx, req.DocumentId, req.Page, req.PageSize, "")
	if err != nil {
		return nil, err
	}

	list := make([]types.KnowledgeDocumentChunkInfo, 0, len(chunks))
	for _, chunk := range chunks {
		var metadata map[string]interface{}
		if chunk.Metadata != "" {
			_ = json.Unmarshal([]byte(chunk.Metadata), &metadata)
		}

		status := 0
		if chunk.Status == "enable" {
			status = 1
		}

		list = append(list, types.KnowledgeDocumentChunkInfo{
			Id:                  chunk.Id,
			KnowledgeBaseId:     chunk.KnowledgeBaseId,
			KnowledgeDocumentId: chunk.KnowledgeDocumentId,
			ChunkText:           chunk.ChunkText,
			ChunkSize:           int(chunk.ChunkSize),
			Metadata:            metadata,
			Status:              status,
			CreatedAt:           chunk.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:           chunk.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetKnowledgeDocumentChunksResp{
		Total: count,
		List:  list,
	}, nil
}
