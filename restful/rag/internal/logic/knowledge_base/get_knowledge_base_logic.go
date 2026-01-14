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

type GetKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库详情
func NewGetKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeBaseLogic {
	return &GetKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeBaseLogic) GetKnowledgeBase(req *types.GetKnowledgeBaseReq) (resp *types.GetKnowledgeBaseResp, err error) {
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

	// Security Check: Ensure KB belongs to current tenant
	if kb.TenantId != tenantId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权访问此知识库")
	}

	return &types.GetKnowledgeBaseResp{
		KnowledgeBaseInfo: types.KnowledgeBaseInfo{
			Id:                     kb.Id,
			TenantId:               kb.TenantId,
			Name:                   kb.Name,
			Avatar:                 kb.Avatar.String,
			Language:               kb.Language,
			Description:            kb.Description.String,
			EmbdId:                 kb.EmbdId,
			Permission:             kb.Permission,
			CreatedBy:              kb.CreatedBy,
			DocNum:                 int64(kb.DocNum), // uint64 -> int64
			TokenNum:               int64(kb.TokenNum),
			ChunkNum:               int64(kb.ChunkNum),
			SimilarityThreshold:    kb.SimilarityThreshold,
			VectorSimilarityWeight: kb.VectorSimilarityWeight,
			Status:                 kb.Status,
			ParserId:               kb.ParserId,
			ParserConfig:           kb.ParserConfig.String,
			CreatedTime:            kb.CreatedTime,
			UpdatedTime:            kb.UpdatedTime,
		},
	}, nil
}
