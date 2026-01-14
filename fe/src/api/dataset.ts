import request from '@/utils/request';
import type { ProcessFilesReq, PreviewChunksReq, PreviewChunksResp } from '@/types';

// 上传文档
export const uploadDataset = async (knowledgeId: string, files: File[]) => {
    const formData = new FormData();

    // 添加 knowledge_base_id 参数
    formData.append('knowledge_base_id', knowledgeId);

    // 添加文件
    files.forEach((file) => {
        formData.append('file', file);
    });

    return request.post(
        `/knowledge_document/upload`,
        formData,
        {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        }
    );
};

// Step 2: Preview Chunks (Real API)
export const previewChunks = (knowledgeId: string, req: PreviewChunksReq) => {
    return request.post<any, PreviewChunksResp>(`/knowledge/${knowledgeId}/document/preview`, req, {
        timeout: 60000, // 60 秒超时，解析大文件可能较慢
    });
};

export const clearDatasetDocuments = (knowledgeId: string) => {
    return request.delete<any, void>(`/knowledge/${knowledgeId}/all`);
};

// Save Chunk Settings & Process Files (Real API)
export const saveChunkSettings = (knowledgeId: string, req: ProcessFilesReq) => {
    return request.post<any, void>(`/knowledge/${knowledgeId}/document/chunk/save`, req);
};

// Alias for backward compatibility
export const processFiles = saveChunkSettings;

// uploadFiles 别名，兼容旧组件
export const uploadFiles = (
    knowledgeId: string,
    files: File[],
    _onProgress?: (progress: number) => void
) => {
    // 忽略 onProgress 参数，因为新API不支持
    return uploadDataset(knowledgeId, files);
};
