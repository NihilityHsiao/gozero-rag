import type { UserInfo, JwtToken } from './index';

export interface LoginResponse {
    token: JwtToken;
    user: UserInfo;
}
