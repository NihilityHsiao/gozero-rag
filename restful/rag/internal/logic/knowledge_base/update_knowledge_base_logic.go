package knowledge_base

import (
	"context"
	"database/sql"
	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新知识库
func NewUpdateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateKnowledgeBaseLogic {
	return &UpdateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateKnowledgeBaseLogic) UpdateKnowledgeBase(req *types.UpdateKnowledgeBaseReq) (resp *types.UpdateKnowledgeBaseResp, err error) {
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == knowledge_base.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
		}
		return nil, xerr.NewInternalErrMsg(err.Error())
	}

	// Security Check
	if kb.TenantId != tenantId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权修改此知识库")
	}

	// Update Fields
	isUpdated := false
	if req.Name != "" {
		kb.Name = req.Name
		isUpdated = true
	}
	if req.Avatar != "" {
		kb.Avatar = sql.NullString{String: req.Avatar, Valid: true}
		isUpdated = true
	}
	if req.Language != "" {
		kb.Language = req.Language
		isUpdated = true
	}
	if req.Description != "" {
		kb.Description = sql.NullString{String: req.Description, Valid: true}
		isUpdated = true
	}
	if req.Permission != "" {
		kb.Permission = req.Permission
		isUpdated = true
	}
	if req.SimilarityThreshold > 0 {
		kb.SimilarityThreshold = req.SimilarityThreshold
		isUpdated = true
	}
	if req.VectorSimilarityWeight > 0 {
		kb.VectorSimilarityWeight = req.VectorSimilarityWeight
		isUpdated = true
	}
	if req.ParserId != "" {
		kb.ParserId = req.ParserId
		isUpdated = true
	}
	if req.ParserConfig != "" {
		kb.ParserConfig = sql.NullString{String: req.ParserConfig, Valid: true}
		isUpdated = true
	}

	if isUpdated {
		err = l.svcCtx.KnowledgeBaseModel.Update(l.ctx, kb)
		if err != nil {
			return nil, xerr.NewInternalErrMsg(err.Error())
		}
	}

	return &types.UpdateKnowledgeBaseResp{}, nil
}
