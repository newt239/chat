import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { api } from "#/lib/api/client";
import { clearAuthAtom } from "#/providers/store/auth";

type UpdatePasswordInput = {
  currentPassword: string;
  newPassword: string;
};

/** パスワードを変更する。変更後はサーバー側の全セッションが失効する */
export const useUpdatePassword = () =>
  useMutation({
    mutationFn: async (input: UpdatePasswordInput) => {
      const { error } = await api.PATCH("/api/users/me/password", { body: input });
      if (error) {
        throw new Error(error.error);
      }
    },
  });

/** アカウントを削除してログイン画面へ戻す */
export const useDeleteAccount = () => {
  const clearAuth = useSetAtom(clearAuthAtom);
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async () => {
      const { error } = await api.DELETE("/api/users/me");
      if (error) {
        throw new Error(error.error);
      }
    },
    onSuccess: async () => {
      clearAuth();
      queryClient.clear();
      await navigate({ to: "/login" });
    },
  });
};
