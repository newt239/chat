import { useMutation } from "@tanstack/react-query";
import { useSetAtom } from "jotai";
import { useNavigate } from "react-router";

import { api } from "#/lib/api/client";
import { paths } from "#/lib/paths";
import { setAuthAtom } from "#/providers/store/auth";

import type { components } from "#/lib/api/schema";

type AuthResponse = components["schemas"]["AuthResponse"];

export const useRegister = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (data: { email: string; password: string; displayName: string }) => {
      const { data: response, error } = await api.POST("/api/auth/register", {
        body: data,
      });
      if (error || !response) {
        throw new Error(error?.error || "登録に失敗しました");
      }
      return response;
    },
    onSuccess: (data: AuthResponse) => {
      setAuth({ user: data.user, accessToken: data.accessToken, refreshToken: data.refreshToken });
      void navigate(paths.app());
    },
  });
};
