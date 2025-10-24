import { useMutation } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import type { components } from "@/lib/api/schema";

import { api } from "@/lib/api/client";
import { setAuthAtom, clearAuthAtom } from "@/lib/store/auth";

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
    onSuccess: (data: components["schemas"]["AuthResponse"]) => {
      setAuth({ user: data.user, accessToken: data.accessToken, refreshToken: data.refreshToken });
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
    onSuccess: (data: components["schemas"]["AuthResponse"]) => {
      setAuth({ user: data.user, accessToken: data.accessToken, refreshToken: data.refreshToken });
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
    },
  });
}
