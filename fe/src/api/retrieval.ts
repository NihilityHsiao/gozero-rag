import request from '@/utils/request';
import type { RetrieveReq, RetrieveResp, GetRetrieveLogResp } from '../types/retrieval';

export const retrievalApi = {
    // 知识库召回测试
    retrieve: async (data: RetrieveReq): Promise<RetrieveResp> => {
        const { knowledge_base_id, ...rest } = data;
        // URL without /v1 prefix as configured in vite proxy
        return request.post<any, RetrieveResp>(`/retrieval/${knowledge_base_id}`, rest);
    },

    // 获取知识库召回记录
    getRetrievalLog: async (knowledgeBaseId: string, params: { page: number; page_size: number }): Promise<GetRetrieveLogResp> => {
        // URL without /v1 prefix as configured in vite proxy
        return request.get<any, GetRetrieveLogResp>(`/retrieval/log/${knowledgeBaseId}`, { params });
    },
};
