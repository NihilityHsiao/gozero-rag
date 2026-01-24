// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

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

type DeleteAllDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除知识库下的所有文档
func NewDeleteAllDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAllDocumentLogic {
	return &DeleteAllDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAllDocumentLogic) DeleteAllDocument(req *types.DeleteAllDocumentReq) (resp *types.DeleteAllDocumentResp, err error) {
	// 1. 获取当前用户ID (也是租户ID)
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 检查知识库权限
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == knowledge_base.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
		}
		return nil, xerr.NewInternalErrMsg("查询知识库失败")
	}

	// 权限校验: 只有 Owner (租户ID == 用户ID) 有权清空
	// 知识库的 TenantId 必须等于当前用户的 ID
	if kb.TenantId != userId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权操作此知识库，仅创建者/管理员可清空")
	}

	// 3. 查找所有非 indexing 状态的文档 (以便删除 MinIO 文件)
	docs, err := l.svcCtx.KnowledgeDocumentModel.FindAllNonIndexingByKbId(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to find docs for deletion: %v", err)
		return nil, xerr.NewInternalErrMsg("查询文档失败")
	}

	if len(docs) == 0 {
		return &types.DeleteAllDocumentResp{}, nil
	}

	// 4. 异步或同步删除 MinIO 文件 & 收集 ID 删除向量
	// 由于数量可能较多，且错误不影响 DB 删除 (脏数据)，我们可以遍历删除
	// TODO: 如果数量非常大，应该放入后台任务。此处假设数量在合理范围内 (<1000)
	var docIds []string
	for _, doc := range docs {
		docIds = append(docIds, doc.Id)
		if doc.StoragePath.Valid && doc.StoragePath.String != "" && doc.SourceType == "local" {
			// 尝试删除 OSS 文件
			// 忽略错误，只记录日志
			err := l.svcCtx.OssClient.RemoveObject(l.ctx, l.svcCtx.Config.Oss.BucketName, doc.StoragePath.String)
			if err != nil {
				l.Errorf("Failed to delete oss object %s: %v", doc.StoragePath.String, err)
			}
		}
	}

	// 5. 删除向量库中的切片
	if len(docIds) > 0 {
		err = l.svcCtx.ChunkModel.DeleteByDocIds(l.ctx, req.Id, docIds)
		if err != nil {
			l.Errorf("Failed to delete chunks from es for kb %s: %v", req.Id, err)
			// 不阻断后续 DB 删除，避免死循环无法清空
		}
	}

	// 6. 数据库删除
	err = l.svcCtx.KnowledgeDocumentModel.DeleteAllNonIndexingByKbId(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to delete docs from db: %v", err)
		return nil, xerr.NewInternalErrMsg("清空文档失败")
	}

	return &types.DeleteAllDocumentResp{}, nil
}
