import { useMutation } from "@tanstack/react-query";
import { useSetAtom } from "jotai";

import type { components } from "@/lib/api/schema";

import { api } from "@/lib/api/client";
import { router } from "@/lib/router";
import { setAuthAtom } from "@/providers/store/auth";

type AuthResponse = components["schemas"]["AuthResponse"];

export function useLogin() {
  const setAuth = useSetAtom(setAuthAtom);

  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      const { data: response, error } = await api.POST("/api/auth/login", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "ログインに失敗しました");
      }
      return response;
    },
    onSuccess: (data: AuthResponse) => {
      setAuth({ user: data.user, accessToken: data.accessToken, refreshToken: data.refreshToken });

      const workspaceStorage = localStorage.getItem("workspace-storage");

      if (workspaceStorage) {
        try {
          const parsed = JSON.parse(workspaceStorage);
          const currentWorkspaceId = parsed.state?.currentWorkspaceId;

          if (currentWorkspaceId) {
            // ワークスペースが選択されている場合はそのページにリダイレクト
            router.navigate({
              to: "/app/$workspaceId",
              params: { workspaceId: currentWorkspaceId },
            });
            return;
          }
        } catch (error) {
          console.warn("ワークスペース情報の解析に失敗しました:", error);
        }
      }

      // ワークスペース情報がない場合は通常のアプリページにリダイレクト
      router.navigate({ to: "/app" });
    },
  });
}
