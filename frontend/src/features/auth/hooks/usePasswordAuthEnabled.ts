import { useQuery } from "@connectrpc/connect-query";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";

/** パスワード認証はサーバーの設定で無効にできるため、フォームを出すかを問い合わせる */
export const usePasswordAuthEnabled = () =>
  useQuery(AuthService.method.getAuthConfig, {}, { select: (res) => res.passwordAuthEnabled })
    .data === true;
