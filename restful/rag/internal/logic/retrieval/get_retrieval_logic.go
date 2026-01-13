// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package retrieval

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gozero-rag/internal/model/knowledge_retrieval_log"
	"gozero-rag/internal/rag_core/retriever"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRetrievalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 知识库召回测试
func NewGetRetrievalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRetrievalLogic {
	return &GetRetrievalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRetrievalLogic) getRetrieveRequestFromReq(req *types.RetrieveReq) (*retriever.RetrieveRequest, error) {
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		return nil, xerr.NewInternalErrMsg("知识库不存在")
	}

	emb, err := l.svcCtx.UserApiModel.FindOne(l.ctx, kb.EmbeddingModelId)
	if err != nil {
		return nil, xerr.NewInternalErrMsg("embedding model not found")
	}

	rnk, err := l.svcCtx.UserApiModel.FindOne(l.ctx, req.RetrievalConfig.HybridStrategy.RerankModelID)
	if err != nil {
		return nil, xerr.NewInternalErrMsg("rerank model not found")
	}

	mode := ""

	hybridType := ""

	switch req.RetrievalMode {
	case retriever.RetrieveModeFulltext:
		mode = retriever.RetrieveModeFulltext
	case retriever.RetrieveModeVector:
		mode = retriever.RetrieveModeVector
	case retriever.RetrieveModeHybrid:
		mode = retriever.RetrieveModeHybrid

		if req.RetrievalConfig.HybridStrategy.Type == retriever.HybridRankTypeWeighted {
			hybridType = retriever.HybridRankTypeWeighted
		} else if req.RetrievalConfig.HybridStrategy.Type == retriever.HybridRankTypeRerank {
			hybridType = retriever.HybridRankTypeRerank
		} else {
			logx.Errorf("invalid hybrid type: %v", req.RetrievalConfig.HybridStrategy.Type)
			return nil, fmt.Errorf("不支持的混合召回类型")
		}

	default:
		logx.Errorf("invalid retrieve type: %v", req.RetrievalMode)
		return nil, fmt.Errorf("不支持的召回模式")
	}

	// check hybrid

	ret := &retriever.RetrieveRequest{
		Query:                req.Query,
		KnowledgeBaseId:      req.KnowledgeBaseId,
		TopK:                 req.RetrievalConfig.TopK,
		EmbeddingModelConfig: retriever.ModelConfig{ModelName: emb.ModelName, BaseUrl: emb.BaseUrl, ApiKey: emb.ApiKey},
		RerankModelConfig:    retriever.ModelConfig{ModelName: rnk.ModelName, BaseUrl: rnk.BaseUrl, ApiKey: rnk.ApiKey},
		Mode:                 mode,
		ScoreThreshold:       req.RetrievalConfig.ScoreThreshold,
		HybridRankType:       hybridType,
		VectorWeight:         req.RetrievalConfig.HybridStrategy.Weights.Vector,
		KeywordWeight:        req.RetrievalConfig.HybridStrategy.Weights.Keyword,
	}

	return ret, nil

}
func (l *GetRetrievalLogic) GetRetrieval(req *types.RetrieveReq) (resp *types.RetrieveResp, err error) {

	start := time.Now()

	retrieveReq, err := l.getRetrieveRequestFromReq(req)
	if err != nil {
		return nil, err
	}

	docs, err := l.svcCtx.RetrieveSvc.Query(l.ctx, retrieveReq)
	if err != nil {
		logx.Errorf("检索失败:%v", err)
		return nil, xerr.NewInternalErrMsg("检索失败")
	}

	chunks := make([]types.RetrievalChunk, 0, len(docs))
	for _, doc := range docs {
		meta := retriever.ExtractDocMeta(doc)
		chunks = append(chunks, types.RetrievalChunk{
			ChunkID: meta.ChunkID,
			DocID:   meta.DocID,
			DocName: "", // 目前 metadata 中没有 doc_name，通常需要单独查询或在索引时冗余存储
			Content: doc.Content,
			Score:   doc.Score(),
			Source:  meta.Type,
		})
	}

	cost := time.Since(start).Milliseconds()
	userId, _ := common.GetUidFromCtx(l.ctx)

	// 记录日志
	go func() {
		// 序列化 RetrievalParams
		paramsBytes, _ := json.Marshal(req.RetrievalConfig)
		logEntry := &knowledge_retrieval_log.KnowledgeRetrievalLog{
			KnowledgeBaseId: req.KnowledgeBaseId,
			UserId:          userId,
			Query:           req.Query,
			RetrievalMode:   req.RetrievalMode,
			RetrievalParams: sql.NullString{
				String: string(paramsBytes),
				Valid:  true,
			},
			ChunkCount: int64(len(chunks)),
			TimeCostMs: cost,
		}

		_, logErr := l.svcCtx.KnowledgeRetrievalLogModel.Insert(context.Background(), logEntry)
		if logErr != nil {
			logx.Errorf("记录召回日志失败, err:%v, log:%v", logErr, logEntry)
		}

	}()

	resp = &types.RetrieveResp{
		KnowledgeBaseID: req.KnowledgeBaseId,
		//DocIDs:          []string{"019b7e12-6c72-7835-9c8c-143ab03aa6e8", "019b7eff-6c72-7835-9c8c-143ab03aa6e8"},
		TimeCostMs: cost,
		Chunks:     chunks,
	}

	return
}
