// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改知识库(描述信息、启用或禁用状态、知识库名称）
func NewUpdateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateKnowledgeBaseLogic {
	return &UpdateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateKnowledgeBaseLogic) UpdateKnowledgeBase(req *types.UpdateKnowledgeBaseReq) (resp *types.UpdateKnowledgeBaseResp, err error) {
	if req.KnowledgeBaseId == 0 || req.Name == "" {
		return nil, xerr.NewBadRequestErrMsg("知识库id和知识库名称不能为空")
	}

	// Fetch existing knowledge base to preserve EmbeddingModelId
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		if err == knowledge.ErrNotFound {
			return nil, xerr.NewErrMsg("知识库不存在")
		}
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, err.Error())
	}

	// Construct model_ids JSON
	modelMap := make(map[string]uint64)
	if req.QaId > 0 {
		modelMap["qa"] = req.QaId
	}
	if req.ChatId > 0 {
		modelMap["chat"] = req.ChatId
	}
	if req.RerankId > 0 {
		modelMap["rerank"] = req.RerankId
	}
	if req.RewriteId > 0 {
		modelMap["rewrite"] = req.RewriteId
	}

	// If request model ids are empty, maybe we should preserve existing ones?
	// The current requirement is just to fix embedding_model_id.
	// We will keep the logic for modelIds as is for now, assuming the frontend sends valid data or full replacement is intended for these specific fields.

	modelIdsJson := []byte("{}")
	if len(modelMap) != 0 {
		modelIdsJson, _ = json.Marshal(modelMap)
	}

	err = l.svcCtx.KnowledgeBaseModel.Update(l.ctx, &knowledge.KnowledgeBase{
		Id:               req.KnowledgeBaseId,
		Name:             req.Name,
		Description:      sql.NullString{String: req.Description, Valid: true},
		Status:           req.Status,
		EmbeddingModelId: kb.EmbeddingModelId, // Preserve existing EmbeddingModelId
		ModelIds:         string(modelIdsJson),
	})

	return &types.UpdateKnowledgeBaseResp{}, err
}
