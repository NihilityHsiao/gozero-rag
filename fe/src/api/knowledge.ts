import request from '@/utils/request';
import type {
  GetKnowledgeBaseListReq,
  GetKnowledgeBaseListResp,
  CreateKnowledgeBaseReq,
  CreateKnowledgeBaseResp,
  UpdateKnowledgeBaseReq,
  KnowledgeBaseInfo,
} from '@/types';

export const getKnowledgeList = (params: GetKnowledgeBaseListReq) => {
  return request.get<any, GetKnowledgeBaseListResp>('/knowledge/list', { params });
};

export const createKnowledgeBase = (data: CreateKnowledgeBaseReq) => {
  return request.post<any, CreateKnowledgeBaseResp>('/knowledge', data);
};

export const getKnowledgeDetail = (id: number) => {
  return request.get<any, KnowledgeBaseInfo>(`/knowledge/${id}`);
};

export const deleteKnowledgeBase = (id: number) => {
  return request.delete<any, void>(`/knowledge/${id}`);
};

export const updateKnowledgeBase = (id: number, data: UpdateKnowledgeBaseReq) => {
  return request.put<any, void>(`/knowledge/${id}`, data);
};
