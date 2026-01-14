package knowledge_base

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/model/tenant_llm"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/google/uuid"

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
	// 1. 获取当前用户和租户
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		// 根据Phase 3设计，知识库属于租户。必须有tenantId。
		return nil, err
	}

	// 2. 检查名称是否重复 (在同一租户下)
	exist, err := l.svcCtx.KnowledgeBaseModel.FindOneByTenantIdName(l.ctx, tenantId, req.Name)
	if err != nil && err != knowledge_base.ErrNotFound {
		return nil, xerr.NewInternalErrMsg("查询知识库失败")
	}
	if exist != nil {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "知识库名称已存在")
	}

	// 3. 校验 embd_id 格式并确认属于当前租户
	// embd_id 格式: 模型名称@厂商
	parts := strings.Split(req.EmbdId, "@")
	if len(parts) != 2 {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "Embedding模型ID格式错误，应为: 模型名称@厂商")
	}
	llmName := parts[0]
	llmFactory := parts[1]

	// 查询 tenant_llm 验证模型归属权
	_, err = l.svcCtx.TenantLlmModel.FindByTenantFactoryName(l.ctx, tenantId, llmFactory, llmName)
	if err != nil {
		if err == tenant_llm.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "该Embedding模型不属于当前租户或不存在")
		}
		l.Errorf("查询租户模型失败: %v", err)
		return nil, xerr.NewInternalErrMsg("验证模型失败")
	}

	// 4. 构造 KnowledgeBase 对象
	// 4. 构造 KnowledgeBase 对象
	kbUuid, err := uuid.NewV7()
	if err != nil {
		return nil, xerr.NewInternalErrMsg("UUID generation failed")
	}
	kbId := kbUuid.String()

	now := time.Now()
	nowUnix := now.UnixMilli()

	kb := &knowledge_base.KnowledgeBase{
		Id:          kbId,
		TenantId:    tenantId,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Avatar:      sql.NullString{String: req.Avatar, Valid: req.Avatar != ""},
		Permission:  req.Permission,
		CreatedBy:   userId,
		EmbdId:      req.EmbdId, // string in types.go

		Status:                 1, // Enabled
		DocNum:                 0,
		TokenNum:               0,
		ChunkNum:               0,
		SimilarityThreshold:    req.SimilarityThreshold,
		VectorSimilarityWeight: req.VectorSimilarityWeight,

		Language:    req.Language,
		CreatedTime: nowUnix,
		UpdatedTime: nowUnix,
		CreatedDate: now,
		UpdatedDate: now,
	}

	// 5. 插入数据库
	_, err = l.svcCtx.KnowledgeBaseModel.Insert(l.ctx, kb)
	if err != nil {
		logx.Errorf("CreateKnowledgeBase Insert error: %v", err)
		return nil, xerr.NewInternalErrMsg("创建知识库失败")
	}

	return &types.CreateKnowledgeBaseResp{
		KnowledgeBaseInfo: types.KnowledgeBaseInfo{
			Id:                     kbId,
			TenantId:               tenantId,
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
		},
	}, nil
}
