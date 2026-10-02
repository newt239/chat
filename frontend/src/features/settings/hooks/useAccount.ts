import { useMutation } from "@connectrpc/connect-query";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { signOut } from "#/lib/session";

/** パスワードを変更する。変更後はサーバー側の全セッションが失効する */
export const useUpdatePassword = () => useMutation(UserService.method.updatePassword);

/** アカウントを削除してログイン画面へ戻す */
export const useDeleteAccount = () =>
  useMutation(UserService.method.deleteMe, { onSuccess: signOut });
