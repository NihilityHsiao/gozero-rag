// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"
	"database/sql"
	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/model/knowledge_document"
	"gozero-rag/internal/mq" // shared mq struct
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchParseDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量解析文档
func NewBatchParseDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchParseDocumentLogic {
	return &BatchParseDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchParseDocumentLogic) BatchParseDocument(req *types.BatchParseDocumentReq) (resp *types.BatchParseDocumentResp, err error) {
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	if len(req.DocumentIds) == 0 {
		return &types.BatchParseDocumentResp{}, nil
	}

	// 1. 验证知识库权限
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		if err == knowledge_base.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
		}
		return nil, xerr.NewInternalErrMsg(err.Error())
	}
	if kb.TenantId != tenantId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权操作此知识库")
	}

	// 2. 查询并验证文档属于该知识库
	docs, err := l.svcCtx.KnowledgeDocumentModel.FindManyByIdsAndKbId(l.ctx, req.DocumentIds, req.KnowledgeBaseId)
	if err != nil {
		logx.Errorf("BatchParseDocument FindMany error: %v", err)
		return nil, xerr.NewInternalErrMsg("查询文档失败")
	}

	if len(docs) == 0 {
		return &types.BatchParseDocumentResp{}, nil
	}

	// 3. 遍历更新状态并发送消息队列
	// 注意：这里简单循环处理，如果量大可以考虑并发或批量Update
	for _, doc := range docs {
		// 更新状态
		doc.RunStatus = knowledge_document.RunStatePending
		doc.Status = 1 // 确保启用
		doc.Progress = 0
		doc.ProgressMsg = sql.NullString{String: "", Valid: true}

		err := l.svcCtx.KnowledgeDocumentModel.Update(l.ctx, doc)
		if err != nil {
			logx.Errorf("BatchParseDocument verify doc ownership or update failed: docId=%s, err=%v", doc.Id, err)
			continue
		}

		// 发送消息
		err = l.svcCtx.MqPusherClient.PublishDocumentIndex(l.ctx, &mq.KnowledgeDocumentIndexMsg{
			UserId:          userId,
			TenantId:        tenantId,
			KnowledgeBaseId: req.KnowledgeBaseId,
			DocumentId:      doc.Id,
		})
		if err != nil {
			logx.Errorf("BatchParseDocument push mq failed: docId=%s, err=%v", doc.Id, err)
			// MQ发送失败是否回滚状态？目前暂不回滚，前端可以看到状态为pending但通过日志排查
			// 或者可以设置 doc.RunStatus = "failed"
		}
	}

	return &types.BatchParseDocumentResp{}, nil
}
