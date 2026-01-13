// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/xerr"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeBaseInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识库详情
func NewGetKnowledgeBaseInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeBaseInfoLogic {
	return &GetKnowledgeBaseInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeBaseInfoLogic) GetKnowledgeBaseInfo(req *types.GetKnowledgeBaseInfoReq) (resp *types.KnowledgeBaseInfo, err error) {
	if req.KnowledgeBaseId == 0 {
		return nil, xerr.NewBadRequestErrMsg("知识库id为空")
	}

	knowledgeBase, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, uint64(req.KnowledgeBaseId))

	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
		}
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "知识库查询失败")
	}

	resp = &types.KnowledgeBaseInfo{
		Id:               int64(knowledgeBase.Id),
		Name:             knowledgeBase.Name,
		Description:      knowledgeBase.Description.String,
		Status:           knowledgeBase.Status,
		EmbeddingModelId: knowledgeBase.EmbeddingModelId,
		CreatedAt:        knowledgeBase.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        knowledgeBase.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// Parse model_ids
	if knowledgeBase.ModelIds != "" {
		var modelIdsMap map[string]uint64
		if err := json.Unmarshal([]byte(knowledgeBase.ModelIds), &modelIdsMap); err == nil {
			var modelIdsInfos []types.KnowledgeBaseModelIdsInfo
			for k, v := range modelIdsMap {
				if v > 0 {
					modelName := ""
					if model, err := l.svcCtx.UserApiModel.FindOne(l.ctx, v); err == nil {
						modelName = model.ConfigName
					} else {
						l.Logger.Errorf("Failed to find model id: %d, err: %v", v, err)
					}

					modelIdsInfos = append(modelIdsInfos, types.KnowledgeBaseModelIdsInfo{
						ModelId:   v,
						ModelType: k,
						ModelName: modelName,
					})
				}
			}
			resp.ModelIds = modelIdsInfos
		} else {
			l.Errorf("Failed to unmarshal model_ids: %v", err)
		}
	}

	return
}
