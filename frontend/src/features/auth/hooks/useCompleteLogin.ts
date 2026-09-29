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

/** ログイン方法によらず、発行されたトークンを保存してアプリへ移動する。参加リンクから来たときはそのワークスペースを開く */
export const useCompleteLogin = (joinedWorkspaceId: string | null) => {
  const setAuth = useSetAtom(setAuthAtom);
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const navigate = useNavigate();

  return ({ accessToken, refreshToken, user }: AuthResponse) => {
    setAuth({ accessToken, refreshToken, user });

    // ワークスペースが選択済みならそのページへ、なければアプリのトップへ
    const workspaceId = joinedWorkspaceId ?? currentWorkspaceId;
    if (workspaceId) {
      void navigate({ params: { workspaceId }, to: "/app/$workspaceId" });
    } else {
      void navigate({ to: "/app" });
    }
  };
};
