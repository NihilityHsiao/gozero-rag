export interface UserInfo {
  user_id: number;
  username: string;
}

export interface JwtToken {
  access_token: string;
  refresh_token: string;
  expire_at: number;
  uid: number;
}

export interface KnowledgeBaseModelIdsInfo {
  model_id: number;
  model_name: string;
  model_type: string;
}

export interface KnowledgeBaseInfo {
  id: number;
  name: string;
  description: string;
  avatar?: string;
  embedding_model?: string;
  status: number; // 0-disabled, 1-enabled
  embedding_model_id: number;
  model_ids: KnowledgeBaseModelIdsInfo[];
  created_at: string;
  updated_at: string;
  qa_model_id?: number;
  chat_model_id?: number;
  rerank_model_id?: number;
  rewrite_model_id?: number;
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
  embedding_id: number;
  rerank_id?: number;
  rewrite_id?: number;
  qa_id?: number;
  chat_id?: number;
}

export interface CreateKnowledgeBaseResp {
  id: number;
}

export interface UpdateKnowledgeBaseReq {
  name?: string;
  description?: string;
  avatar?: string;
  status?: number;
  qa_model_id?: number;
  chat_model_id?: number;
  embedding_model_id?: number;
  rerank_model_id?: number;
  rewrite_model_id?: number;
}

// Document related types
export type DocStatus = 'disable' | 'pending' | 'indexing' | 'enable' | 'fail';

export interface KnowledgeDocumentInfo {
  id: number;
  knowledge_base_id: number;
  doc_name: string;
  doc_type: string; // pdf/word/txt
  doc_size: number;
  description: string;
  status: DocStatus; // disable, pending, enable, fail
  chunk_count: number;
  parser_config?: SegmentationSettings;
  err_msg: string;
  created_at: string;
  updated_at: string;
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
  knowledge_base_id: number;
  knowledge_document_id: string;
  chunk_text: string;
  chunk_size: number;
  metadata: ChunkMetadata;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface GetKnowledgeDocumentChunksReq {
  knowledge_base_id: number;
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
  knowledge_base_id: number;
  doc_id: string;
}

export interface UserApiInfo {
  id: number;
  user_id: number;
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
