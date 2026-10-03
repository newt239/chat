import { useMutation } from "@connectrpc/connect-query";

import { usePushNotifications } from "#/features/notification/hooks/usePushNotifications";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { signOut } from "#/lib/session";

/** 別のユーザーがこの端末でログインしても通知が届かないよう、認証が有効なうちにプッシュ通知の登録を外す */
export const useDisablePushBeforeSignOut = () => {
  const push = usePushNotifications();

  return async () => {
    try {
      await push.disable();
    } catch (error) {
      console.warn("プッシュ通知の登録を解除できませんでした", error);
    }
  };
};

/** サーバー側のセッションを失効させてからローカルの認証情報を破棄する */
export const useLogout = () =>
  useMutation(AuthService.method.logout, {
    onMutate: useDisablePushBeforeSignOut(),
    // 失効に失敗してもローカルからは必ずログアウトする
    onSettled: signOut,
  });
