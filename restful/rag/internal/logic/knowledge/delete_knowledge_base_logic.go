// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"fmt"

	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/xerr"
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
	kbId := uint64(req.KnowledgeBaseId)

	// 1. 检查知识库是否存在
	_, err = l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, kbId)
	if err != nil {
		if err == knowledge.ErrNotFound {
			return nil, xerr.NewErrCode(xerr.KnowledgeBaseNotFoundError)
		}
		l.Logger.Errorf("查询知识库失败: %v", err)
		return nil, xerr.NewInternalErrMsg("查询知识库失败")
	}

	// 2. 检查是否有正在索引中的文档（不允许删除）
	docs, err := l.svcCtx.KnowledgeDocumentModel.FindListByKbId(l.ctx, kbId, "")
	if err != nil {
		l.Logger.Errorf("查询文档列表失败: %v", err)
		return nil, xerr.NewInternalErrMsg("查询文档列表失败")
	}

	for _, doc := range docs {
		if doc.Status == knowledge.StatusDocumentIndexing {
			return nil, xerr.NewBadRequestErrMsg("存在正在索引中的文档，无法删除知识库")
		}
	}

	// 3. 删除 Milvus Collection (TODO: VectorClient 未注入 ServiceContext，暂时跳过)
	// collectionName := fmt.Sprintf("kb_%d", kbId)
	// if err := l.svcCtx.VectorClient.DropCollection(l.ctx, collectionName); err != nil {
	// 	l.Logger.Errorf("删除 Milvus Collection 失败: %v", err)
	// }
	_ = fmt.Sprintf("kb_%d", kbId) // Placeholder

	// 4. 删除知识库记录（文档和 Chunk 通过外键级联删除或手动删除）
	// 注意：这里简化处理，假设没有外键约束，需要手动删除
	if err := l.svcCtx.KnowledgeBaseModel.Delete(l.ctx, kbId); err != nil {
		l.Logger.Errorf("删除知识库失败: %v", err)
		return nil, xerr.NewInternalErrMsg("删除知识库失败")
	}

	l.Logger.Infof("知识库 %d 删除成功", kbId)
	return &types.DeleteKnowledgeBaseResp{}, nil
}
