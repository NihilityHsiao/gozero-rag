// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"encoding/json"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/mq"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveChunkSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存分片配置
func NewSaveChunkSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveChunkSettingsLogic {
	return &SaveChunkSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveChunkSettingsLogic) SaveChunkSettings(req *types.SaveChunkSettingsReq) (resp *types.SaveChunkSettingsResp, err error) {
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	jsonStr, err := json.Marshal(req.Settings)
	if err != nil {
		logx.Errorw("json marshal error", logx.Field("detail", err.Error()))
		return nil, xerr.NewInternalErrMsg("json marshal error")
	}

	// 等待处理
	err = l.svcCtx.KnowledgeDocumentModel.UpdateParserConfigAndStatus(l.ctx, req.FileIds, string(jsonStr), knowledge.StatusDocumentDisable, knowledge.StatusDocumentPending)
	if err != nil {
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "update parser config error")
	}

	// 注意: 只有 disable状态 -> pending状态才能发消息队列
	// 7. 发送Kafka消息,触发异步文档分片处理, 交给 knowledge_index去做

	for _, fileId := range req.FileIds {
		// 8. 记录成功指标
		msg := &mq.KnowledgeDocumentIndexMsg{
			UserId:          userId,
			KnowledgeBaseId: req.KnowledgeBaseId,
			DocumentId:      fileId,
		}
		err = l.svcCtx.MqPusherClient.PublishDocumentIndex(l.ctx, msg)
		if err != nil {
			logx.Errorf("发送消息失败:%v", err)
			return nil, xerr.NewErrCodeMsg(xerr.InternalError, "send document process to kafka error")
		}

		logx.Infof("发送doc process到Kafka消息队列: %v", msg)
	}

	return &types.SaveChunkSettingsResp{}, nil
}
