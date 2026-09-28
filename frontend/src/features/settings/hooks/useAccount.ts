import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { clearAuthAtom } from "#/providers/store/auth";

/** パスワードを変更する。変更後はサーバー側の全セッションが失効する */
export const useUpdatePassword = () => useMutation(UserService.method.updatePassword);

/** アカウントを削除してログイン画面へ戻す */
export const useDeleteAccount = () => {
  const clearAuth = useSetAtom(clearAuthAtom);
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation(UserService.method.deleteMe, {
    onSuccess: async () => {
      clearAuth();
      queryClient.clear();
      await navigate({ to: "/login" });
    },
  });
};
