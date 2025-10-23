import { useMutation } from "@tanstack/react-query";
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
        throw new Error((error as any)?.error || "Login failed");
      }
      return response as any;
    },
    onSuccess: (data: any) => {
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
        throw new Error((error as any)?.error || "Registration failed");
      }
      return response as any;
    },
    onSuccess: (data: any) => {
      setAuth(data.user, data.accessToken, data.refreshToken);
    },
  });
}

export function useLogout() {
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const refreshToken = useAuthStore((state) => state.refreshToken);

  return useMutation({
    mutationFn: async () => {
      if (refreshToken) {
        await apiClient.POST("/api/auth/logout", {
          body: { refreshToken } as any,
        });
      }
    },
    onSuccess: () => {
      clearAuth();
    },
  });
}
