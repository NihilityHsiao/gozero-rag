import request from '@/utils/request';
import type { GetUserApiListReq, GetUserApiListResp, AddUserApiReq, AddUserApiResp } from '@/types';

/**
 * 获取用户模型配置列表
 * @param user_id 用户ID
 * @param params 查询参数（model_type, status）
 */
export const getUserApiList = (
    user_id: number,
    params?: GetUserApiListReq
) => {
    return request.get<any, GetUserApiListResp>(`/user/api/${user_id}`, { params });
};

/**
 * 添加用户模型配置
 */
export const addUserApi = (data: AddUserApiReq) => {
    return request.post<any, AddUserApiResp>('/user/api', data);
};

/**
 * 删除用户模型配置
 */
export const deleteUserApi = (id: number) => {
    return request.delete<any, any>(`/user/api/${id}`);
};

/**
 * 设置默认模型
 */
export const setUserModelDefault = (data: { user_id: number; model_id: number; model_type: string }) => {
    return request.post<any, any>('/user/api/set_default', data);
};
