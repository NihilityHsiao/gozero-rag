// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateKnowledgeBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建知识库
func NewCreateKnowledgeBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateKnowledgeBaseLogic {
	return &CreateKnowledgeBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateKnowledgeBaseLogic) CreateKnowledgeBase(req *types.CreateKnowledgeBaseReq) (resp *types.CreateKnowledgeBaseResp, err error) {
	if req.Name == "" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "知识库名称不能为空")
	}

	uid, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. 检查是否存在同名知识库
	_, err = l.svcCtx.KnowledgeBaseModel.FindOneByName(l.ctx, req.Name)
	if err == nil {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "知识库已存在")
	}
	if !errors.Is(err, knowledge.ErrNotFound) {
		logx.Errorf("查找知识库失败,err:%v, req:%v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "创建知识库失败")
	}

	// 2. 校验 Embedding Model
	if req.EmbeddingId == 0 {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "请选择Embedding模型")
	}

	embeddingModel, err := l.svcCtx.UserApiModel.FindOne(l.ctx, req.EmbeddingId)
	if err != nil {
		if errors.Is(err, user_api.ErrNotFound) {
			return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "Embedding模型不存在")
		}
		l.Errorf("查询Embedding模型失败: %v, id: %d", err, req.EmbeddingId)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}
	if embeddingModel.UserId != uid {
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "无权使用该Embedding模型")
	}
	if embeddingModel.ModelType != "embedding" {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "所选模型不是Embedding类型")
	}

	// 3. 构建 model_ids JSON
	modelIds := map[string]uint64{
		"rerank":  req.RerankId,
		"rewrite": req.RewriteId,
		"qa":      req.QaId,
		"chat":    req.ChatId,
	}

	// 可选：校验其他模型ID (略，如果这里校验会更严谨，但为了性能暂时只校验核心的Embedding)
	// 如果需要校验，可以批量查询或者单独查询。

	modelIdsJson, err := json.Marshal(modelIds)
	if err != nil {
		logx.Errorf("序列化model_ids失败: %v", err)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	// 4. 插入数据库
	ret, err := l.svcCtx.KnowledgeBaseModel.Insert(l.ctx, &knowledge.KnowledgeBase{
		Name:             req.Name,
		Description:      sql.NullString{String: req.Description, Valid: true},
		Status:           1,
		EmbeddingModelId: req.EmbeddingId,
		ModelIds:         string(modelIdsJson),
	})

	if err != nil {
		logx.Errorf("创建知识库失败,err:%v, req:%v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "创建知识库失败")
	}

	kbId, err := ret.LastInsertId()
	if err != nil {
		logx.Errorf("创建知识库,获取id失败,err:%v, req:%v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "创建知识库失败")
	}

	return &types.CreateKnowledgeBaseResp{Id: kbId}, nil
}
