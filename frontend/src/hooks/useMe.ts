import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useAtomValue } from "jotai";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { sessionAtom } from "#/providers/store/auth";

/** ログイン中ユーザーのプロフィール。ログイン前に呼んでも取得しない */
export const useMe = () => {
  const hasSession = useAtomValue(sessionAtom) !== null;
  return useQuery(UserService.method.getMe, hasSession ? {} : skipToken, {
    select: (res) => res.user,
  });
};
