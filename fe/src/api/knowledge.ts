import request from '@/utils/request';
import type {
  GetKnowledgeBaseListReq,
  GetKnowledgeBaseListResp,
  CreateKnowledgeBaseReq,
  CreateKnowledgeBaseResp,
  UpdateKnowledgeBaseReq,
  KnowledgeBaseInfo,
  GraphDetailResp,

} from '@/types';

export const getKnowledgeList = (params: GetKnowledgeBaseListReq) => {
  return request.get<any, GetKnowledgeBaseListResp>('/knowledge/list', { params });
};

export const createKnowledgeBase = (data: CreateKnowledgeBaseReq) => {
  return request.post<any, CreateKnowledgeBaseResp>('/knowledge/create', data);
};

export const getKnowledgeDetail = (id: string) => {
  return request.get<any, KnowledgeBaseInfo>(`/knowledge/${id}`);
};

export const deleteKnowledgeBase = (id: string) => {
  return request.delete<any, void>(`/knowledge/${id}`);
};

export const updateKnowledgeBase = (id: string, data: UpdateKnowledgeBaseReq) => {
  return request.put<any, void>(`/knowledge/${id}`, data);
};

// 获取租户 LLM 列表（用于创建知识库时选择 Embedding 模型）
export const getTenantLlmList = (params?: { model_type?: string }) => {
  return request.get<any, { list: any[] }>('/tenant/llm', { params });
};

// 更新知识库权限
export const updateKnowledgeBasePermission = (id: string, permission: string) => {
  return request.patch<any, void>(`/knowledge/${id}/permission`, { permission });
};

// 获取知识图谱数据
export const getKnowledgeGraph = (kb_id: string, params?: { limit?: number }) => {
  return request.get<any, GraphDetailResp>(`/knowledge/${kb_id}/graph`, { params });
};

// 搜索图谱节点
export const searchKnowledgeGraph = (kb_id: string, params: { q: string }) => {
  return request.get<any, GraphDetailResp>(`/knowledge/${kb_id}/graph/search`, { params });
};
