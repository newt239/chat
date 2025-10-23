import { useMutation } from "@tanstack/react-query";

import type { components } from "@/lib/api/schema";

import { apiClient } from "@/lib/api/client";
import { useAuthStore } from "@/lib/store/auth";

export function useLogin() {
  const setAuth = useAuthStore((state) => state.setAuth);

  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      const { data: response, error } = await apiClient.POST("/api/auth/login", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "Login failed");
      }
      return response;
    },
    onSuccess: (data: components["schemas"]["AuthResponse"]) => {
      setAuth(data.user, data.accessToken, data.refreshToken);
    },
  });
}

export function useRegister() {
  const setAuth = useAuthStore((state) => state.setAuth);

  return useMutation({
    mutationFn: async (data: { email: string; password: string; displayName: string }) => {
      const { data: response, error } = await apiClient.POST("/api/auth/register", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "Registration failed");
      }
      return response;
    },
    onSuccess: (data: components["schemas"]["AuthResponse"]) => {
      setAuth(data.user, data.accessToken, data.refreshToken);
    },
  });
}

export function useLogout() {
  const clearAuth = useAuthStore((state) => state.clearAuth);

  return useMutation({
    mutationFn: async () => {
      await apiClient.POST("/api/auth/logout", {});
    },
    onSuccess: () => {
      clearAuth();
    },
  });
}
