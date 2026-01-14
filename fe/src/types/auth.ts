import type { UserInfo, JwtToken, TenantInfo } from './index';

export interface LoginRequest {
    email: string;
    password: string;
}

export interface RegisterRequest {
    nickname: string;
    email: string;
    password: string;
    confirm_password: string;
}

export interface LoginResponse {
    token: JwtToken;
    user: UserInfo;
    current_tenant: TenantInfo;
    tenants: TenantInfo[];
}
