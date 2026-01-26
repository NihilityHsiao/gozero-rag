// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package graph

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeGraphLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取知识图谱数据
func NewGetKnowledgeGraphLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeGraphLogic {
	return &GetKnowledgeGraphLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeGraphLogic) GetKnowledgeGraph(req *types.GraphReq) (resp *types.GraphDetailResp, err error) {
	// 默认 limit
	limit := req.Limit
	if limit <= 0 {
		limit = 500
	}

	// 1. 调用 Model 获取图谱数据
	entities, relations, err := l.svcCtx.NebulaGraphModel.GetGraph(l.ctx, req.KbId, limit)
	if err != nil {
		l.Errorw("get graph failed", logx.Field("err", err))
		return nil, err
	}

	// 2. 转换为 API 响应格式
	nodes := make([]types.GraphNode, 0, len(entities))
	links := make([]types.GraphLink, 0, len(relations))

	// 简单的度数统计 (可选，作为 val)
	degreeMap := make(map[string]int)
	for _, r := range relations {
		degreeMap[r.SrcId]++
		degreeMap[r.DstId]++
	}

	for _, e := range entities {
		val := degreeMap[e.Name]
		if val == 0 {
			val = 1
		}
		nodes = append(nodes, types.GraphNode{
			Id:          e.Name,
			Name:        e.Name,
			Type:        e.Type,
			Description: e.Description,
			Val:         val,
			SourceId:    e.SourceId,
		})
	}

	for _, r := range relations {
		links = append(links, types.GraphLink{
			Source:      r.SrcId,
			Target:      r.DstId,
			Description: r.Description,
			Weight:      r.Weight,
		})
	}

	return &types.GraphDetailResp{
		Nodes: nodes,
		Links: links,
	}, nil
}
