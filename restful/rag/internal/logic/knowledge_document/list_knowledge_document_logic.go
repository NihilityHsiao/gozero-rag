// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListKnowledgeDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文档列表
func NewListKnowledgeDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListKnowledgeDocumentLogic {
	return &ListKnowledgeDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListKnowledgeDocumentLogic) ListKnowledgeDocument(req *types.ListKnowledgeDocumentReq) (resp *types.ListKnowledgeDocumentResp, err error) {
	// 获取当前用户ID
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询知识库信息
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		l.Errorf("查询知识库失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
	}

	// 多租户权限校验：
	// 1. 查询用户关联的所有租户
	tenantIds, err := l.svcCtx.UserTenantModel.FindTenantsByUserId(l.ctx, userId)
	if err != nil {
		l.Errorf("查询用户租户失败: %v", err)
		return nil, xerr.NewInternalErrMsg("查询权限失败")
	}

	// 2. 检查知识库所在租户是否在用户的租户列表中
	hasAccess := false
	for _, tid := range tenantIds {
		if tid == kb.TenantId {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权限访问该知识库")
	}

	// 3. 进一步检查权限：如果是 me 权限且不是创建者，拒绝访问
	if kb.Permission == "me" && kb.CreatedBy != userId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权限访问该知识库")
	}

	// 查询文档列表
	docs, err := l.svcCtx.KnowledgeDocumentModel.FindListByKnowledgeBaseId(
		l.ctx,
		req.KnowledgeBaseId,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		l.Errorf("查询文档列表失败: %v", err)
		return nil, xerr.NewInternalErrMsg("查询文档列表失败")
	}

	// 统计总数
	total, err := l.svcCtx.KnowledgeDocumentModel.CountByKnowledgeBaseId(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		l.Errorf("统计文档数量失败: %v", err)
		return nil, xerr.NewInternalErrMsg("统计文档数量失败")
	}

	// 转换为响应类型
	list := make([]types.KnowledgeDocumentInfo, 0, len(docs))
	for _, doc := range docs {
		list = append(list, types.KnowledgeDocumentInfo{
			Id:              doc.Id,
			KnowledgeBaseId: doc.KnowledgeBaseId,
			DocName:         doc.DocName.String,
			DocType:         doc.DocType,
			DocSize:         doc.DocSize,
			StoragePath:     doc.StoragePath.String,
			Description:     doc.Description.String,
			Status:          doc.Status,
			RunStatus:       doc.RunStatus,
			ChunkNum:        doc.ChunkNum,
			TokenNum:        doc.TokenNum,
			ParserConfig:    doc.ParserConfig,
			Progress:        doc.Progress,
			ProgressMsg:     doc.ProgressMsg.String,
			CreatedBy:       doc.CreatedBy,
			CreatedTime:     doc.CreatedTime,
			UpdatedTime:     doc.UpdatedTime,
		})
	}

	return &types.ListKnowledgeDocumentResp{
		Total: total,
		List:  list,
	}, nil
}
