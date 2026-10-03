import { useMutation } from "@connectrpc/connect-query";

import { usePushNotifications } from "#/features/settings/hooks/usePushNotifications";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { signOut } from "#/lib/session";

/** サーバー側のセッションを失効させてからローカルの認証情報を破棄する */
export const useLogout = () => {
  const push = usePushNotifications();

  return useMutation(AuthService.method.logout, {
    // 別のユーザーがこの端末でログインしても通知が届かないよう、認証が有効なうちに外す
    onMutate: async () => {
      try {
        await push.disable();
      } catch (error) {
        console.warn("プッシュ通知の登録を解除できませんでした", error);
      }
    },
    // 失効に失敗してもローカルからは必ずログアウトする
    onSettled: signOut,
  });
};
