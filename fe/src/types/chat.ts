// 检索设置配置
export interface ChatRetrieveConfig {
    mode: 'vector' | 'fulltext' | 'hybrid';
    top_k: number;
    score: number; // 阈值
    rerank_mode?: 'weighted' | 'rerank';
    rerank_model_id?: number;
    rerank_vector_weight?: number;
    rerank_keyword_weight?: number;
}

export interface ChatModelConfig {
    model_id?: number;
    // Extra fields to support textual config storage
    model_name?: string;
    model_factory?: string;

    knowledge_base_ids?: string[];
    system_prompt?: string;
    temperature?: number;

    // Flattened Retrieval Settings for internal usage (optional compat)
    retrieval_mode?: 'vector' | 'fulltext' | 'hybrid';
    top_k?: number;
    score_threshold?: number;
    hybrid_strategy_type?: 'weighted' | 'rerank';
    weight_vector?: number;
    weight_keyword?: number;
    rerank_model_id?: number;
    // Extra fields for rerank
    rerank_model_name?: string;
    rerank_model_factory?: string;
}

export interface Conversation {
    id: string;
    title: string;
    message_count: number;
    updated_at: string;
    created_at: string;

    // Optional details that might be loaded later or not present in list view
    user_id?: number;
    model_config?: ChatModelConfig;
    status?: 1 | 2 | 3; // 1-normal, 2-archived, 3-deleted
}

// Alias for backwards compatibility if needed, or just replace usage
export type ChatConversation = Conversation;

export interface ChatMessage {
    id: string;
    conversation_id?: string;
    seq_id?: number;
    role: 'user' | 'assistant' | 'system' | 'tool';
    content: string;

    type: 'text' | 'json';
    token_count: number;
    created_at: string;

    // UI Extended Fields
    isError?: boolean;
    errorMsg?: string;

    extra?: {
        citations?: ChatRetrievalChunk[];
        reasoning?: string;
        usage?: {
            prompt_tokens: number;
            completion_tokens: number;
            total_tokens: number;
        };
    };
}

// 开启新对话请求
export interface StartNewChatReq {
    llm_id: string; // 模型名称@厂商
    enable_quote_doc?: boolean;
    enable_llm_keyword_extract?: boolean;
    enable_tts?: boolean;
    system_prompt?: string;
    kb_ids?: string[];
    temperature?: number;

    retrieval_config?: {
        mode?: string;
        rerank_mode?: string;
        rerank_vector_weight?: number;
        top_n?: number;
        rerank_id?: string; // 模型名称@厂商
        top_k?: number;
        score?: number;
    };
}

// 开启新对话响应
export interface StartNewChatResp {
    conversation_id: string;
}

// SSE 请求参数
export interface ChatReq {
    conversation_id: string;
    message: string;
    chat_model_id: number;
    prompt: string;
    temperature: number;
    knowledge_base_ids: string[];
    chat_retrieve_config: ChatRetrieveConfig;
}

// 引用来源块
export interface ChatRetrievalChunk {
    chunk_id: string;
    doc_id: string;
    doc_name: string;
    content: string;
    score: number;
    source: string;
}

// SSE 响应结构 (流式返回)
// SSE 响应结构 (流式返回)
export interface ChatResp {
    msg_id: string;
    type: 'text' | 'citation' | 'reasoning' | 'tool_use' | 'finish' | 'error';

    // Optional Fields based on type
    content?: string;                  // type='text'
    reasoning_content?: string;        // type='reasoning'
    tool_call_id?: string;             // type='tool_use'
    retrieval_docs?: ChatRetrievalChunk[]; // type='citation'
    token_usage?: number;              // type='finish'
    finish_reason?: string;            // type='finish'
    error_msg?: string;                // type='error'
}

// --- New Endpoints Types ---

// Re-export or alias if we want to keep terminology distinct, but user asked to perfect list. 
// Let's reuse Conversation for list items as per API.
export type ChatListInfo = Conversation;

export interface GetConversationListReq {
    page: number;
    page_size: number;
}

export interface GetConversationListResp {
    list: Conversation[];
    total: number;
}

export interface UpdateConversationReq {
    conversation_id: string;
    title: string;
}
