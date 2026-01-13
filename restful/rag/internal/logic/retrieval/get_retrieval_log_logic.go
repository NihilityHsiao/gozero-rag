// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package retrieval

import (
	"context"
	"encoding/json"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRetrievalLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 知识库召回记录
func NewGetRetrievalLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRetrievalLogLogic {
	return &GetRetrievalLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRetrievalLogLogic) GetRetrievalLog(req *types.GetRetrieveLogReq) (resp *types.GetRetrieveLogResp, err error) {
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	// Calculate Offset
	offset := (req.Page - 1) * req.PageSize
	if offset < 0 {
		offset = 0
	}

	// 1. Get List with Limit/Offset
	logs, err := l.svcCtx.KnowledgeRetrievalLogModel.FindLogList(l.ctx, req.KnowledgeBaseId, uint64(userId), offset, req.PageSize)
	if err != nil {
		l.Logger.Errorf("Failed to get retrieval logs: %v", err)
		return nil, xerr.NewInternalErrMsg("获取召回记录失败")
	}

	// 2. Count Total
	total, err := l.svcCtx.KnowledgeRetrievalLogModel.CountLog(l.ctx, req.KnowledgeBaseId, uint64(userId))
	if err != nil {
		l.Logger.Errorf("Failed to count retrieval logs: %v", err)
	}

	// 3. Convert to Response Types
	respLogs := make([]types.RetrieveLog, 0, len(logs))
	for _, log := range logs {
		var params types.RetrievalConfig
		if log.RetrievalParams.Valid {
			if err := json.Unmarshal([]byte(log.RetrievalParams.String), &params); err != nil {
				l.Logger.Errorf("Failed to unmarshal retrieval params: %v", err)
			}
		}

		respLogs = append(respLogs, types.RetrieveLog{
			Id:              log.Id,
			KnowledgeBaseId: log.KnowledgeBaseId,
			Query:           log.Query,
			RetrievalMode:   log.RetrievalMode,
			RetrievalParams: params,
			ChunkCount:      int(log.ChunkCount),
			TimeCostMs:      log.TimeCostMs,
			CreatedAt:       log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	resp = &types.GetRetrieveLogResp{
		Total: total,
		Logs:  respLogs,
	}

	return
}
