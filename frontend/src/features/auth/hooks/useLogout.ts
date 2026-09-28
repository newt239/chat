import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { clearAuthAtom } from "#/providers/store/auth";

/** サーバー側のセッションを失効させてからローカルの認証情報を破棄する */
export const useLogout = () => {
  const clearAuth = useSetAtom(clearAuthAtom);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation(AuthService.method.logout, {
    // 失効に失敗してもローカルからは必ずログアウトする
    onSettled: async () => {
      clearAuth();
      queryClient.clear();
      await navigate({ to: "/login" });
    },
  });
};
