import { useNavigate } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { setAuthAtom } from "#/providers/store/auth";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

import type { User } from "#/gen/chat/v1/user_pb";

type AuthResponse = {
  accessToken: string;
  refreshToken: string;
  user?: User;
};

/** ログイン方法によらず、発行されたトークンを保存してアプリへ移動する */
export const useCompleteLogin = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const navigate = useNavigate();

  return ({ accessToken, refreshToken, user }: AuthResponse) => {
    setAuth({ accessToken, refreshToken, user });

    // ワークスペースが選択済みならそのページへ、なければアプリのトップへ
    if (currentWorkspaceId) {
      void navigate({ params: { workspaceId: currentWorkspaceId }, to: "/app/$workspaceId" });
    } else {
      void navigate({ to: "/app" });
    }
  };
};
