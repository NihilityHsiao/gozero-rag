import request from '@/utils/request';
import type {
  GetKnowledgeDocumentListReq,
  GetKnowledgeDocumentListResp,
  KnowledgeDocumentInfo,
  GetKnowledgeDocumentChunksReq,
  GetKnowledgeDocumentChunksResp,
} from '@/types';

export const getDocDetail = (docId: string) => {
  return request.get<any, KnowledgeDocumentInfo>(`/knowledge_document/${docId}`);
};

export const getDocumentChunks = (params: GetKnowledgeDocumentChunksReq) => {
  return request.get<any, GetKnowledgeDocumentChunksResp>(`/knowledge_document/${params.document_id}/chunks`, { params });
};

export const getDocumentList = (knowledgeBaseId: string, params: GetKnowledgeDocumentListReq) => {
  return request.get<any, GetKnowledgeDocumentListResp>(`/knowledge_document`, {
    params: {
      ...params,
      knowledge_base_id: knowledgeBaseId
    }
  });
};

export const batchParseDocument = (knowledgeBaseId: string, documentIds: string[]) => {
  return request.post<any, any>(`/knowledge_document/batch_parse`, {
    knowledge_base_id: knowledgeBaseId,
    document_ids: documentIds,
  });
};
