// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"strconv"
	"time"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取会话列表
func NewGetConversationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationListLogic {
	return &GetConversationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationListLogic) GetConversationList(req *types.GetConversationListReq) (resp *types.GetConversationListResp, err error) {
	userId, _ := common.GetUidFromCtx(l.ctx)
	list, total, err := l.svcCtx.ChatConversationModel.FindListByUserId(l.ctx, strconv.FormatInt(userId, 10), req.Page, req.PageSize)
	if err != nil {
		l.Logger.Errorf("FindListByUserId error: %v, userId: %s", err, userId)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "Failed to get conversation list")
	}

	respList := make([]types.Conversation, 0, len(list))
	for _, item := range list {
		respList = append(respList, types.Conversation{
			Id:           item.Id,
			Title:        item.Title,
			MessageCount: int(item.MessageCount),
			UpdatedAt:    item.UpdatedAt.Format(time.RFC3339),
			CreatedAt:    item.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.GetConversationListResp{
		List:  respList,
		Total: total,
	}, nil
}
