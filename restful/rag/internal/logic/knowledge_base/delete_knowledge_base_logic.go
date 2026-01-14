package knowledge_base

import (
	"context"
	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除知识库
func NewDeleteKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteKnowledgeBaseLogic {
	return &DeleteKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteKnowledgeBaseLogic) DeleteKnowledgeBase(req *types.DeleteKnowledgeBaseReq) (resp *types.DeleteKnowledgeBaseResp, err error) {
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. Fetch KB
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == knowledge_base.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
		}
		return nil, xerr.NewInternalErrMsg(err.Error())
	}

	// 2. Security Check
	if kb.TenantId != tenantId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权删除此知识库")
	}

	// 3. Delete Chunks from Vector DB (Elasticsearch)
	err = l.svcCtx.ChunkModel.DeleteByKbId(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("DeleteByKbId ChunkModel error: %v", err)
		// Proceed? Or Fail? Usually we should try to cleanup or background it.
		// For now, fail hard to avoid inconsistencies, or just log error.
		return nil, xerr.NewInternalErrMsg("删除向量数据失败")
	}

	// 4. Delete Knowledge Documents
	err = l.svcCtx.KnowledgeDocumentModel.DeleteByKbId(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("DeleteByKbId KnowledgeDocumentModel error: %v", err)
		return nil, xerr.NewInternalErrMsg("删除文档数据失败")
	}

	// 5. Delete Knowledge Base
	err = l.svcCtx.KnowledgeBaseModel.Delete(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("Delete KnowledgeBaseModel error: %v", err)
		return nil, xerr.NewInternalErrMsg("删除知识库失败")
	}

	return &types.DeleteKnowledgeBaseResp{}, nil
}
