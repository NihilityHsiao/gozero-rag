// 用户基本信息
export interface UserInfo {
  user_id: string;      // UUID v7
  nickname: string;
  email: string;
  avatar?: string;
  language?: string;
}

// JWT Token
export interface JwtToken {
  access_token: string;     // 核心访问令牌
  refresh_token: string;    // 刷新令牌
  expire_at: number;        // access token过期时间
}

// 租户基本信息
export interface TenantInfo {
  tenant_id: string;
  name: string;
  role: string;             // owner|admin|member
}

export interface KnowledgeBaseModelIdsInfo {
  model_id: number;
  model_name: string;
  model_type: string;
}

export interface KnowledgeBaseInfo {
  id: string;
  name: string;
  description: string;
  avatar?: string;
  embedding_model?: string;
  embd_id: string;
  status: number; // 0-disabled, 1-enabled
  embedding_model_id: number;
  model_ids: KnowledgeBaseModelIdsInfo[];
  permission: string; // me | team
  created_by: string; // 创建者用户ID
  created_at: string;
  updated_at: string;
  qa_model_id?: number;
  chat_model_id?: number;
  rerank_model_id?: number;
  rewrite_model_id?: number;
  parser_id: string;
  parser_config: string;
  similarity_threshold: number;
  vector_similarity_weight: number;
}

export interface GetKnowledgeBaseListReq {
  page?: number;
  page_size?: number;
  status?: number;
}

export interface GetKnowledgeBaseListResp {
  total: number;
  list: KnowledgeBaseInfo[];
}

export interface CreateKnowledgeBaseReq {
  name: string;
  description?: string;
  embd_id: string; // 格式: 模型名称@厂商
}

export interface CreateKnowledgeBaseResp {
  id: string;
}

export interface UpdateKnowledgeBaseReq {
  name?: string;
  description?: string;
  avatar?: string;
  status?: number;
  permission?: string;
  qa_model_id?: number;
  chat_model_id?: number;
  embedding_model_id?: number;
  rerank_model_id?: number;
  rewrite_model_id?: number;
  parser_id?: string;
  parser_config?: string; // JSON string
  similarity_threshold?: number;
  vector_similarity_weight?: number;
}

// Parser ID Enum
export type ParserId = 'general' | 'resume';

// General Parser Config
export interface GeneralParserConfig {
  chunk_token_num: number;
  chunk_overlap_token_num: number;
  separator: string[];
  layout_recognize: boolean;
  qa_num?: number;
  qa_llm_id?: string;
  pdf_parser: 'pdfcpu' | 'eino' | 'deepdoc';
}

// Resume Parser Config
export interface ResumeParserConfig {
  pdf_parser: 'pdfcpu' | 'eino' | 'deepdoc';
}

// Unified Config Type
export type ParserConfig = GeneralParserConfig | ResumeParserConfig;

// Document related types
// run_status 状态类型
export type RunStatus = 'pending' | 'indexing' | 'success' | 'failed' | 'canceled' | 'paused';

export interface KnowledgeDocumentInfo {
  id: string;
  knowledge_base_id: string;
  doc_name: string;
  doc_type: string; // pdf/word/txt/md
  doc_size: number;
  storage_path: string;
  description: string;
  status: number; // 1-启用, 0-禁用
  run_status: RunStatus; // 文档处理状态
  chunk_num: number; // 切片数量
  token_num: number; // token数量
  parser_config: string; // JSON string
  progress: number; // 处理进度 0-100
  progress_msg: string; // 进度信息
  created_by: string;
  created_time: number; // 时间戳(毫秒)
  updated_time: number; // 时间戳(毫秒)
}

export interface GetKnowledgeDocumentListReq {
  page?: number;
  page_size?: number;
  status?: number;
}

export interface GetKnowledgeDocumentListResp {
  total: number;
  list: KnowledgeDocumentInfo[];
}
export interface GetKnowledgeDocumentListResp {
  total: number;
  list: KnowledgeDocumentInfo[];
}

export interface QaPair {
  question: string;
  answer: string;
}

export interface ChunkMetadata {
  h1?: string;
  h2?: string;
  h3?: string;
  _source?: string;
  qa_pairs?: QaPair[];
  _extension?: string;
  _file_name?: string;
  quality_score?: number;
}

export interface KnowledgeDocumentChunkInfo {
  id: string;
  knowledge_base_id: string;
  knowledge_document_id: string;
  chunk_text: string;
  chunk_size: number;
  metadata: ChunkMetadata;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface GetKnowledgeDocumentChunksReq {
  knowledge_base_id: string;
  document_id: string;
  page?: number;
  page_size?: number;
  keyword?: string;
}

export interface GetKnowledgeDocumentChunksResp {
  total: number;
  list: KnowledgeDocumentChunkInfo[];
}

export interface GetDocByDocIdReq {
  knowledge_base_id: string;
  doc_id: string;
}

export interface UserApiInfo {
  id: number;
  user_id: string;
  config_name: string;
  api_key: string;
  base_url: string;
  model_name: string;
  model_type: string;
  model_dim: number;
  max_tokens: number;
  temperature: number;
  top_p: number;
  timeout: number;
  status: number;
  is_default: number;
  created_at: string;
  updated_at: string;
  provider: string;
  icon: string;
}

export interface GetUserApiListReq {
  model_type?: string;
  status?: number;
  page?: number;
  page_size?: number;
}

export interface GetUserApiListResp {
  total: number;
  list: UserApiInfo[];
}

export interface AddUserApiReq {
  config_name: string;
  api_key: string;
  base_url: string;
  model_name: string;
  model_type: string;
  model_dim?: number;
  max_tokens?: number;
  temperature?: number;
  top_p?: number;
  timeout?: number;
  status?: number;
  is_default?: number;
}

export interface AddUserApiResp {
  id: number;
}

export interface DeleteUserApiReq {
  id: number;
}

// Dataset / File Upload types
export interface UploadMultiFileResp {
  file_ids: string[];
}

export interface ChunkPreview {
  index: string;
  content: string;
  length: number;
}

export interface PreviewDocResult {
  doc_id: string;
  doc_name: string;
  total_chunks: number;
  chunks: ChunkPreview[];
  error_msg: string;
}

export interface SegmentationSettings {
  separators: string[];
  max_chunk_length: number;
  chunk_overlap: number;
  pre_clean_rule: {
    clean_whitespace: boolean;
    remove_urls_emails: boolean;
  };
  enable_qa_generation: boolean;
}

export interface PreviewChunksReq {
  doc_id: string;
  settings: SegmentationSettings;
}

export interface PreviewChunksResp extends PreviewDocResult { }

export interface ProcessFilesReq {
  file_ids: string[];
  settings: SegmentationSettings;
}

export interface ProcessFilesResp {
  job_id: string;
}

// LLM 相关类型
export * from './llm';
