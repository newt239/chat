import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { usePushNotifications } from "#/features/settings/hooks/usePushNotifications";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { logger } from "#/lib/logger";
import { clearAuthAtom } from "#/providers/store/auth";

/** サーバー側のセッションを失効させてからローカルの認証情報を破棄する */
export const useLogout = () => {
  const clearAuth = useSetAtom(clearAuthAtom);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const push = usePushNotifications();

  return useMutation(AuthService.method.logout, {
    // 別のユーザーがこの端末でログインしても通知が届かないよう、認証が有効なうちに外す
    onMutate: async () => {
      try {
        await push.disable();
      } catch (error) {
        logger.warn("プッシュ通知の登録を解除できませんでした", error);
      }
    },
    // 失効に失敗してもローカルからは必ずログアウトする
    onSettled: async () => {
      clearAuth();
      queryClient.clear();
      await navigate({ to: "/login" });
    },
  });
};
