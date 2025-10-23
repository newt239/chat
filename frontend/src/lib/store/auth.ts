import { create } from "zustand";
import { persist } from "zustand/middleware";

import { navigateToAppWithWorkspace, navigateToLogin } from "../navigation";

import type { components } from "@/lib/api/schema";

type User = components["schemas"]["User"];

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  setAuth: (user: User, accessToken: string, refreshToken: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      setAuth: (user, accessToken, refreshToken) => {
        localStorage.setItem("accessToken", accessToken);
        localStorage.setItem("refreshToken", refreshToken);
        set({ user, accessToken, refreshToken, isAuthenticated: true });
        navigateToAppWithWorkspace();
      },
      clearAuth: () => {
        localStorage.removeItem("accessToken");
        localStorage.removeItem("refreshToken");
        set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false });
        navigateToLogin();
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // ストレージから復元時にトークンを同期
          const accessToken = localStorage.getItem("accessToken");
          const refreshToken = localStorage.getItem("refreshToken");

          console.log("AuthStore - ストレージ復元:", {
            hasAccessToken: !!accessToken,
            hasRefreshToken: !!refreshToken,
            currentUser: state.user,
            currentIsAuthenticated: state.isAuthenticated,
          });

          if (accessToken && refreshToken) {
            state.accessToken = accessToken;
            state.refreshToken = refreshToken;
            console.log("AuthStore - トークン復元成功");
          } else {
            // トークンが存在しない場合は認証状態をリセット
            console.log("AuthStore - トークンなし、認証状態をリセット");
            state.user = null;
            state.accessToken = null;
            state.refreshToken = null;
            state.isAuthenticated = false;
          }
        }
      },
    }
  )
);
