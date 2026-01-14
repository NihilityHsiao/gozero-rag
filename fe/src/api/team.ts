import request from '@/utils/request';

// ============== 类型定义 ==============

// 团队成员信息
export interface TeamMember {
    user_id: string;
    nickname: string;
    email: string;
    role: string; // owner | admin | member
    status: number;
    joined_time: number;
}

// 我加入的团队信息
export interface JoinedTeam {
    tenant_id: string;
    tenant_name: string;
    owner_name: string;
    role: string;
    joined_time: number;
}

// 邀请响应
export interface CreateInviteResp {
    invite_code: string;
    invite_link: string;
}

// 验证邀请码响应
export interface VerifyInviteResp {
    tenant_id: string;
    tenant_name: string;
    inviter: string;
}

// ============== API 请求 ==============

// 获取当前团队成员列表
export const listMembers = () => {
    return request.get<any, { list: TeamMember[] }>('/team/members');
};

// 获取我加入的团队列表
export const listJoinedTeams = () => {
    return request.get<any, { list: JoinedTeam[] }>('/team/joined');
};

// 发起邀请
export const createInvite = (email: string) => {
    return request.post<any, CreateInviteResp>('/team/invite', { email });
};

// 验证邀请码
export const verifyInvite = (code: string) => {
    return request.get<any, VerifyInviteResp>(`/team/invite/${code}`);
};

// 加入团队
export const joinTeam = (inviteCode: string) => {
    return request.post<any, { success: boolean }>('/team/join', { invite_code: inviteCode });
};
