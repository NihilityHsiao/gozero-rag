export interface HybridWeights {
    vector?: number; // default 0.7
    keyword?: number; // default 0.3
}

export interface HybridStrategy {
    type?: 'weighted' | 'rerank'; // default weighted
    weights?: HybridWeights;
    rerank_model_id?: number;
}

export interface RetrievalConfig {
    top_k?: number; // default 10
    score_threshold?: number; // default 0.6
    hybrid_strategy?: HybridStrategy;
}

export interface RetrieveReq {
    knowledge_base_id: number;
    query: string;
    retrieval_mode?: 'vector' | 'fulltext' | 'hybrid'; // default hybrid
    retrieval_config?: RetrievalConfig;
    doc_ids?: string[];
}

export interface RetrievalChunk {
    chunk_id: string;
    doc_id: string;
    doc_name: string;
    content: string;
    score: number;
    source: string; // vector, keyword
}

export interface RetrieveResp {
    knowledge_base_id: number;
    doc_ids: string[];
    time_cost_ms: number;
    chunks: RetrievalChunk[];
}

export interface RetrieveLog {
    id: number;
    knowledge_base_id: number;
    query: string;
    retrieval_mode: string;
    retrieval_params: RetrievalConfig;
    chunk_count: number;
    time_cost_ms: number;
    created_at: string;
}

export interface GetRetrieveLogReq {
    knowledge_base_id: number;
    page: number;
    page_size: number;
}

export interface GetRetrieveLogResp {
    total: number;
    logs: RetrieveLog[];
}
