import { useQuery } from "@connectrpc/connect-query";

import { UserService } from "#/gen/chat/v1/user_service_pb";

/** ログイン中ユーザーのプロフィールを取得する */
export const useMe = () => useQuery(UserService.method.getMe, {}, { select: (res) => res.user });
