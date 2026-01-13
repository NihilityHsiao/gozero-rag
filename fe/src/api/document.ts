import request from '@/utils/request';
import type {
  GetKnowledgeDocumentListReq,
  GetKnowledgeDocumentListResp,
  KnowledgeDocumentInfo,
  GetKnowledgeDocumentChunksReq,
  GetKnowledgeDocumentChunksResp,
} from '@/types';

export const getDocDetail = (knowledgeBaseId: number, docId: string) => {
  return request.get<any, KnowledgeDocumentInfo>(`/knowledge/${knowledgeBaseId}/get_document/${docId}`);
};

export const getDocumentChunks = (params: GetKnowledgeDocumentChunksReq) => {
  return request.get<any, GetKnowledgeDocumentChunksResp>(`/knowledge/${params.knowledge_base_id}/${params.document_id}/chunks`, { params });
};


export const getDocumentList = (knowledgeBaseId: number, params: GetKnowledgeDocumentListReq) => {
  return request.get<any, GetKnowledgeDocumentListResp>(`/knowledge/${knowledgeBaseId}/list`, { params });
};
