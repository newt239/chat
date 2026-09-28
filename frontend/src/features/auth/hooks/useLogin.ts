import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { api } from "#/lib/api/client";
import { setAuthAtom } from "#/providers/store/auth";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

import type { components } from "#/lib/api/schema";

type AuthResponse = components["schemas"]["AuthResponse"];

export const useLogin = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      const { data: response, error } = await api.POST("/api/auth/login", {
        body: data,
      });
      if (error) {
        throw new Error(error.error);
      }
      return response;
    },
    onSuccess: (data: AuthResponse) => {
      setAuth({ accessToken: data.accessToken, refreshToken: data.refreshToken, user: data.user });

      // ワークスペースが選択済みならそのページへ、なければアプリのトップへ
      if (currentWorkspaceId) {
        void navigate({ params: { workspaceId: currentWorkspaceId }, to: "/app/$workspaceId" });
      } else {
        void navigate({ to: "/app" });
      }
    },
  });
};
