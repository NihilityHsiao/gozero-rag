// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"os"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAllDocsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除知识库的所有文档
func NewDeleteAllDocsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAllDocsLogic {
	return &DeleteAllDocsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAllDocsLogic) DeleteAllDocs(req *types.DeleteAllDocumentReq) (resp *types.DeleteAllDocumentResp, err error) {
	// 1. 先将非 indexing 状态的文档状态置为 disable，防止并发被消费
	err = l.svcCtx.KnowledgeDocumentModel.UpdateStatusByKbId(l.ctx, req.KnowledgeBaseId, "disable", "indexing")
	if err != nil {
		l.Logger.Errorf("UpdateStatusByKbId error: %v, kbId: %d", err, req.KnowledgeBaseId)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "更新文档状态失败")
	}

	// 2. 查询该知识库下所有非 indexing 状态的文档 (此时理论上大部分应该是 disable)
	docs, err := l.svcCtx.KnowledgeDocumentModel.FindListByKbId(l.ctx, req.KnowledgeBaseId, "indexing")
	if err != nil {
		l.Logger.Errorf("FindListByKbId error: %v, kbId: %d", err, req.KnowledgeBaseId)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "查询文档失败")
	}

	for _, doc := range docs {
		// 3. 删除物理文件
		if doc.StoragePath != "" {
			err := os.Remove(doc.StoragePath)
			if err != nil && !os.IsNotExist(err) {
				// 记录日志但不中断流程
				l.Logger.Errorf("Remove file error: %v, path: %s", err, doc.StoragePath)
			}
		}

		// 4. 删除数据库记录
		err := l.svcCtx.KnowledgeDocumentModel.Delete(l.ctx, doc.Id)
		if err != nil {
			l.Logger.Errorf("Delete document error: %v, docId: %s", err, doc.Id)
			// 继续删除下一个
		}
	}

	return &types.DeleteAllDocumentResp{}, nil
}
