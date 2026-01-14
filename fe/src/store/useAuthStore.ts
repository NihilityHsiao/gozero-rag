import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { UserInfo, TenantInfo } from '@/types';
import type { LoginResponse } from '@/types/auth';

interface AuthState {
  token: string | null;
  userInfo: UserInfo | null;
  currentTenant: TenantInfo | null;
  tenants: TenantInfo[];
  isAuthenticated: boolean;
  login: (data: LoginResponse) => void;
  logout: () => void;
  switchTenant: (tenant: TenantInfo) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userInfo: null,
      currentTenant: null,
      tenants: [],
      isAuthenticated: false,
      login: (data) => {
        set({
          token: data.token.access_token,
          userInfo: data.user,
          currentTenant: data.current_tenant,
          tenants: data.tenants,
          isAuthenticated: true,
        });
        localStorage.setItem('token', data.token.access_token);
      },
      logout: () => {
        set({
          token: null,
          userInfo: null,
          currentTenant: null,
          tenants: [],
          isAuthenticated: false,
        });
        localStorage.removeItem('token');
      },
      switchTenant: (tenant) => {
        set({
          currentTenant: tenant,
        });
        // 注意: 切换租户后可能需要重新获取 token
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        token: state.token,
        userInfo: state.userInfo,
        currentTenant: state.currentTenant,
        tenants: state.tenants,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
