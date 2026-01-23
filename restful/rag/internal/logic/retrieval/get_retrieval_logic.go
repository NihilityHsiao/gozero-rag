package retrieval

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gozero-rag/internal/model/knowledge_retrieval_log"
	"gozero-rag/internal/rag_core/retriever"
	"gozero-rag/internal/tools/llmx"
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
	// 1. 获取知识库
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		return nil, xerr.NewInternalErrMsg("知识库不存在")
	}

	// 2. 获取 Embedding 模型配置 (TenantLlmModel)
	embModelName, embFactory := llmx.GetModelNameFactory(kb.EmbdId)
	embLlm, err := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(l.ctx, kb.TenantId, embFactory, embModelName)
	if err != nil {
		logx.Errorf("Get embedding model failed: tenantId=%s, factory=%s, model=%s, err=%v", kb.TenantId, embFactory, embModelName, err)
		return nil, xerr.NewInternalErrMsg(fmt.Sprintf("Embedding 模型配置不存在: %s", kb.EmbdId))
	}

	// 3. 获取 Rerank 模型配置 (TenantLlmModel)
	var rnkConfig retriever.ModelConfig
	if req.RetrievalConfig.HybridStrategy.RerankModelID != "" {
		rnkModelName, rnkFactory := llmx.GetModelNameFactory(req.RetrievalConfig.HybridStrategy.RerankModelID)
		rnkLlm, err := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(l.ctx, kb.TenantId, rnkFactory, rnkModelName)
		if err != nil {
			logx.Errorf("Get rerank model failed: tenantId=%s, factory=%s, model=%s, err=%v", kb.TenantId, rnkFactory, rnkModelName, err)
			return nil, xerr.NewInternalErrMsg(fmt.Sprintf("Rerank 模型配置不存在: %s", req.RetrievalConfig.HybridStrategy.RerankModelID))
		}
		rnkConfig = retriever.ModelConfig{
			ModelName: rnkLlm.LlmName,
			BaseUrl:   rnkLlm.ApiBase.String,
			ApiKey:    rnkLlm.ApiKey.String,
		}
	}

	// 4. 确定召回模式
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
			// 默认 weighted
			hybridType = retriever.HybridRankTypeWeighted
		}
	default:
		// 默认混合
		mode = retriever.RetrieveModeHybrid
		hybridType = retriever.HybridRankTypeWeighted
	}

	ret := &retriever.RetrieveRequest{
		Query:           req.Query,
		KnowledgeBaseId: req.KnowledgeBaseId,
		TopK:            req.RetrievalConfig.TopK,
		EmbeddingModelConfig: retriever.ModelConfig{
			ModelName: embLlm.LlmName,
			BaseUrl:   embLlm.ApiBase.String,
			ApiKey:    embLlm.ApiKey.String,
		},
		RerankModelConfig: rnkConfig,
		Mode:              mode,
		ScoreThreshold:    req.RetrievalConfig.ScoreThreshold,
		HybridRankType:    hybridType,
		VectorWeight:      req.RetrievalConfig.HybridStrategy.Weights.Vector,
		KeywordWeight:     req.RetrievalConfig.HybridStrategy.Weights.Keyword,
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
		return nil, xerr.NewInternalErrMsg(fmt.Sprintf("检索失败: %v", err))
	}

	chunks := make([]types.RetrievalChunk, 0, len(docs))
	for _, doc := range docs {
		meta := retriever.ExtractDocMeta(doc)
		chunks = append(chunks, types.RetrievalChunk{
			ChunkID: meta.ChunkID,
			DocID:   meta.DocID,
			DocName: "", // ES 中如果未存 doc_name，则为空
			Content: doc.Content,
			Score:   doc.Score(),
			Source:  meta.Type,
		})
	}

	cost := time.Since(start).Milliseconds()
	userId, _ := common.GetUidFromCtx(l.ctx)

	// 记录日志 (异步)
	go func() {
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
			logx.Errorf("记录召回日志失败, err:%v", logErr)
		}
	}()

	resp = &types.RetrieveResp{
		KnowledgeBaseID: req.KnowledgeBaseId,
		DocIDs:          req.DocIDs,
		TimeCostMs:      cost,
		Chunks:          chunks,
	}

	return resp, nil
}
