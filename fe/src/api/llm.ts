import request from '@/utils/request';
import type {
    LlmFactoryInfo,
    TenantLlmInfo,
    TenantLlmGroupByFactory,
    AddTenantLlmRequest,
    AddTenantLlmResponse,
    UpdateTenantLlmRequest,
} from '@/types';

/**
 * 获取系统支持的 LLM 厂商列表 (需要认证)
 */
export const listLlmFactories = () => {
    return request.get<any, { list: LlmFactoryInfo[] }>('/llm/factories');
};

/**
 * 批量添加租户 LLM 配置
 */
export const addTenantLlm = (data: AddTenantLlmRequest) => {
    return request.post<any, AddTenantLlmResponse>('/tenant/llm', data);
};

/**
 * 获取租户 LLM 配置列表
 */
export const listTenantLlm = (params?: {
    llm_factory?: string;
    model_type?: string;
    status?: number;
    page?: number;
    page_size?: number;
}) => {
    return request.get<any, { total: number; list: TenantLlmInfo[] }>('/tenant/llm', { params });
};

/**
 * 获取租户 LLM 配置列表 (按厂商分组)
 */
export const listTenantLlmGrouped = () => {
    return request.get<any, { list: TenantLlmGroupByFactory[] }>('/tenant/llm/grouped');
};

/**
 * 更新租户 LLM 配置
 */
export const updateTenantLlm = (id: number, data: UpdateTenantLlmRequest) => {
    return request.put<any, void>(`/tenant/llm/${id}`, data);
};

/**
 * 删除租户 LLM 配置
 */
export const deleteTenantLlm = (id: number) => {
    return request.delete<any, void>(`/tenant/llm/${id}`);
};
