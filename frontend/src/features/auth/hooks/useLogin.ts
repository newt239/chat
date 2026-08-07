import { useMutation } from "@tanstack/react-query";
import { useSetAtom } from "jotai";
import { useNavigate } from "react-router";

import { api } from "#/lib/api/client";
import { paths } from "#/lib/paths";
import { setAuthAtom } from "#/providers/store/auth";

import type { components } from "#/lib/api/schema";

type AuthResponse = components["schemas"]["AuthResponse"];

export const useLogin = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const navigate = useNavigate();

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
            void navigate(paths.workspace(currentWorkspaceId));
            return;
          }
        } catch (error) {
          console.warn("ワークスペース情報の解析に失敗しました:", error);
        }
      }

      // ワークスペース情報がない場合は通常のアプリページにリダイレクト
      void navigate(paths.app());
    },
  });
};
