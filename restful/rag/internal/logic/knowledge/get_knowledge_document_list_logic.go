// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"gozero-rag/internal/xerr"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeDocumentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据知识库id,获取知识库的文档列表
func NewGetKnowledgeDocumentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeDocumentListLogic {
	return &GetKnowledgeDocumentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeDocumentListLogic) GetKnowledgeDocumentList(req *types.GetKnowledgeDocumentListReq) (resp *types.GetKnowledgeDocumentListResp, err error) {

	// todo: 暂时不查status, 现在status是int,但数据库存的是string
	// 后面要将api里的status改为string
	status := ""
	count, err := l.svcCtx.KnowledgeDocumentModel.Count(l.ctx, status)
	if err != nil {
		logx.Errorf("GetKnowledgeDocumentList count error: %v,req:%v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "获取知识库文档列表失败")
	}

	list, err := l.svcCtx.KnowledgeDocumentModel.FindList(l.ctx, req.KnowledgeBaseId, req.Page, req.PageSize, status)
	if err != nil {
		logx.Errorf("GetKnowledgeDocumentList find list error: %v,req:%v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "获取知识库文档列表失败")
	}

	respList := make([]types.KnowledgeDocumentInfo, 0, len(list))
	for _, v := range list {
		respList = append(respList, types.KnowledgeDocumentInfo{
			Id:              v.Id,
			KnowledgeBaseId: v.KnowledgeBaseId,
			DocName:         v.DocName,
			DocType:         v.DocType,
			DocSize:         v.DocSize,
			Description:     v.Description.String,
			Status:          v.Status,
			ChunkCount:      v.ChunkCount,
			ErrMsg:          v.ErrMsg,
		})
	}

	return &types.GetKnowledgeDocumentListResp{
		List:  respList,
		Total: count,
	}, nil
}
