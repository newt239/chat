import { useMutation } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import type { components } from "@/lib/api/schema";

import { api } from "@/lib/api/client";
import { navigateToAppWithWorkspace, navigateToLogin } from "@/lib/navigation";
import { setAuthAtom, clearAuthAtom } from "@/providers/store/auth";

type AuthResponse = components["schemas"]["AuthResponse"];

export function useLogin() {
  const setAuth = useSetAtom(setAuthAtom);

  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      const { data: response, error } = await api.POST("/api/auth/login", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "Login failed");
      }
      return response;
    },
    onSuccess: (data: AuthResponse) => {
      setAuth({ user: data.user, accessToken: data.accessToken, refreshToken: data.refreshToken });
      navigateToAppWithWorkspace();
    },
  });
}

export function useRegister() {
  const setAuth = useSetAtom(setAuthAtom);

  return useMutation({
    mutationFn: async (data: { email: string; password: string; displayName: string }) => {
      const { data: response, error } = await api.POST("/api/auth/register", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "Registration failed");
      }
      return response;
    },
    onSuccess: (data: AuthResponse) => {
      setAuth({ user: data.user, accessToken: data.accessToken, refreshToken: data.refreshToken });
      navigateToAppWithWorkspace();
    },
  });
}

export function useLogout() {
  const clearAuth = useSetAtom(clearAuthAtom);

  return useMutation({
    mutationFn: async () => {
      await api.POST("/api/auth/logout", {});
    },
    onSuccess: () => {
      clearAuth();
      navigateToLogin();
    },
  });
}
