import request from '@/utils/request';
import type { LoginResponse } from '@/types/auth';

export interface LoginReq {
  username: string;
  password?: string;
  email?: string;
}

export const login = (data: LoginReq) => {
  return request.post<any, LoginResponse>('/user/login', data);
};

