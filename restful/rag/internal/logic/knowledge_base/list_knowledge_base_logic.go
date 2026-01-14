package knowledge_base

import (
	"context"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库列表
func NewListKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKnowledgeBaseLogic {
	return &ListKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListKnowledgeBaseLogic) ListKnowledgeBase(req *types.ListKnowledgeBaseReq) (resp *types.ListKnowledgeBaseResp, err error) {
	// 获取当前用户ID
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. 查询用户关联的所有租户ID
	tenantIds, err := l.svcCtx.UserTenantModel.FindTenantsByUserId(l.ctx, userId)
	if err != nil {
		l.Errorf("查询用户租户失败: %v", err)
		return nil, xerr.NewInternalErrMsg("查询租户失败")
	}

	if len(tenantIds) == 0 {
		// 用户未关联任何租户，返回空列表
		return &types.ListKnowledgeBaseResp{
			Total: 0,
			List:  []types.KnowledgeBaseInfo{},
		}, nil
	}

	// 2. 多租户查询（SQL中完成所有过滤逻辑）
	list, err := l.svcCtx.KnowledgeBaseModel.FindListByMultiTenants(
		l.ctx,
		userId,
		tenantIds,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		l.Errorf("查询知识库列表失败: %v", err)
		return nil, xerr.NewInternalErrMsg("查询知识库列表失败")
	}

	// 3. 统计总数
	total, err := l.svcCtx.KnowledgeBaseModel.CountByMultiTenants(l.ctx, userId, tenantIds)
	if err != nil {
		l.Errorf("统计知识库数量失败: %v", err)
		return nil, xerr.NewInternalErrMsg("统计失败")
	}

	// 4. 转换为响应（无需应用层过滤，SQL已完成）
	respList := make([]types.KnowledgeBaseInfo, 0, len(list))
	for _, kb := range list {
		info := types.KnowledgeBaseInfo{
			Id:                     kb.Id,
			TenantId:               kb.TenantId,
			Name:                   kb.Name,
			Avatar:                 kb.Avatar.String,
			Language:               kb.Language,
			Description:            kb.Description.String,
			EmbdId:                 kb.EmbdId,
			Permission:             kb.Permission,
			CreatedBy:              kb.CreatedBy,
			DocNum:                 int64(kb.DocNum),
			TokenNum:               int64(kb.TokenNum),
			ChunkNum:               int64(kb.ChunkNum),
			SimilarityThreshold:    kb.SimilarityThreshold,
			VectorSimilarityWeight: kb.VectorSimilarityWeight,
			Status:                 kb.Status,
			CreatedTime:            kb.CreatedTime,
			UpdatedTime:            kb.UpdatedTime,
		}
		respList = append(respList, info)
	}

	return &types.ListKnowledgeBaseResp{
		Total: total,
		List:  respList,
	}, nil
}
