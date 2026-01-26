// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package graph

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchKnowledgeGraphLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 搜索图谱节点
func NewSearchKnowledgeGraphLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchKnowledgeGraphLogic {
	return &SearchKnowledgeGraphLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchKnowledgeGraphLogic) SearchKnowledgeGraph(req *types.GraphSearchReq) (resp *types.GraphDetailResp, err error) {
	if req.Query == "" {
		return &types.GraphDetailResp{
			Nodes: []types.GraphNode{},
			Links: []types.GraphLink{},
		}, nil
	}

	// 1. 调用 Model 搜索节点
	entities, err := l.svcCtx.NebulaGraphModel.SearchGraphNodes(l.ctx, req.KbId, req.Query)
	if err != nil {
		l.Errorw("search graph nodes failed", logx.Field("err", err))
		return nil, err
	}

	// 2. 转换为 API 响应格式
	nodes := make([]types.GraphNode, 0, len(entities))
	for _, e := range entities {
		nodes = append(nodes, types.GraphNode{
			Id:          e.Name,
			Name:        e.Name,
			Type:        e.Type,
			Description: e.Description,
			Val:         5, // 搜索结果稍微突出显示
			SourceId:    e.SourceId,
		})
	}

	return &types.GraphDetailResp{
		Nodes: nodes,
		Links: []types.GraphLink{}, // 搜索只返回节点，暂不返回关系（或可以查找这些节点的1跳关系）
	}, nil
}
