import request from '@/utils/request';
import type { UploadMultiFileResp, ProcessFilesReq, PreviewChunksReq, PreviewChunksResp } from '@/types';

// Step 1: Upload Files
export const uploadFiles = (knowledgeId: number, files: File[], onProgress?: (progress: number) => void) => {
    const formData = new FormData();
    files.forEach((file) => {
        formData.append('files', file);
    });

    return request.post<any, UploadMultiFileResp>(
        `/knowledge/${knowledgeId}/uploadmulti`,
        formData,
        {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
            onUploadProgress: (progressEvent) => {
                if (onProgress && progressEvent.total) {
                    const percentCompleted = Math.round(
                        (progressEvent.loaded * 100) / progressEvent.total
                    );
                    onProgress(percentCompleted);
                }
            },
        }
    );
};

// Step 2: Preview Chunks (Real API)
export const previewChunks = (knowledgeId: number, req: PreviewChunksReq) => {
    return request.post<any, PreviewChunksResp>(`/knowledge/${knowledgeId}/document/preview`, req, {
        timeout: 60000, // 60 秒超时，解析大文件可能较慢
    });
};

export const clearDatasetDocuments = (knowledgeId: number) => {
    return request.delete<any, void>(`/knowledge/${knowledgeId}/all`);
};

// Save Chunk Settings & Process Files (Real API)
export const saveChunkSettings = (knowledgeId: number, req: ProcessFilesReq) => {
    return request.post<any, void>(`/knowledge/${knowledgeId}/document/chunk/save`, req);
};

// Alias for backward compatibility
export const processFiles = saveChunkSettings;
