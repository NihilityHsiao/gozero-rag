// LLM 厂商和租户 LLM 配置相关类型定义

/**
 * LLM 厂商信息
 */
export interface LlmFactoryInfo {
    name: string;
    logo: string;
    tags: string;        // 逗号分隔的模型类型
    tag_list: string[];  // 拆分后的类型数组
    rank: number;
    status: number;
}

/**
 * 租户 LLM 配置信息
 */
export interface TenantLlmInfo {
    id: number;
    tenant_id: string;
    llm_factory: string;
    model_type: string;
    llm_name: string;
    api_key: string;     // 脱敏后的 API Key
    api_base: string;
    max_tokens: number;
    used_tokens: number;
    status: number;
    created_time: number;
    updated_time: number;
}

/**
 * 批量添加模型的单个模型配置
 */
export interface ModelConfig {
    model_type: string;
    llm_name: string;
    max_tokens?: number;
}

/**
 * 批量添加租户 LLM 配置请求
 */
export interface AddTenantLlmRequest {
    llm_factory: string;
    api_key: string;
    api_base: string;
    models: ModelConfig[];
}

/**
 * 批量添加租户 LLM 配置响应
 */
export interface AddTenantLlmResponse {
    success_count: number;
    failed_count: number;
    failed_models: string[];
}

/**
 * 按厂商分组的模型列表
 */
export interface TenantLlmGroupByFactory {
    llm_factory: string;
    factory_logo: string;
    api_base: string;
    models: TenantLlmInfo[];
}

/**
 * 获取厂商列表响应
 */
export interface ListLlmFactoriesResponse {
    list: LlmFactoryInfo[];
}

/**
 * 获取租户 LLM 配置列表响应
 */
export interface ListTenantLlmResponse {
    total: number;
    list: TenantLlmInfo[];
}

/**
 * 获取按厂商分组的列表响应
 */
export interface ListTenantLlmGroupedResponse {
    list: TenantLlmGroupByFactory[];
}

/**
 * 更新租户 LLM 配置请求
 */
export interface UpdateTenantLlmRequest {
    api_key?: string;
    api_base?: string;
    max_tokens?: number;
    status?: number;
}
