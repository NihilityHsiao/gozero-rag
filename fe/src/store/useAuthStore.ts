import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { UserInfo } from '@/types';
import type { LoginResponse } from '@/types/auth';

interface AuthState {
  token: string | null;
  userInfo: UserInfo | null;
  isAuthenticated: boolean;
  login: (data: LoginResponse) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      userInfo: null,
      isAuthenticated: false,
      login: (data) => {
        set({
          token: data.token.access_token,
          userInfo: data.user,
          isAuthenticated: true,
        });
        localStorage.setItem('token', data.token.access_token);
      },
      logout: () => {
        set({
          token: null,
          userInfo: null,
          isAuthenticated: false,
        });
        localStorage.removeItem('token');
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        token: state.token,
        userInfo: state.userInfo,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
