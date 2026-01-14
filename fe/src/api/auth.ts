import request from '@/utils/request';
import type { LoginResponse, LoginRequest, RegisterRequest } from '@/types/auth';

// 登录
export const login = (data: LoginRequest) => {
  return request.post<any, LoginResponse>('/user/login', data);
};

// 注册
export const register = (data: RegisterRequest) => {
  return request.post<any, LoginResponse>('/user/register', data);
};
