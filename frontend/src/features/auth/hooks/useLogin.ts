import { useMutation } from "@connectrpc/connect-query";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { setAuthAtom } from "#/providers/store/auth";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

export const useLogin = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const navigate = useNavigate();

  return useMutation(AuthService.method.login, {
    onSuccess: ({ accessToken, refreshToken, user }) => {
      setAuth({ accessToken, refreshToken, user });

      // ワークスペースが選択済みならそのページへ、なければアプリのトップへ
      if (currentWorkspaceId) {
        void navigate({ params: { workspaceId: currentWorkspaceId }, to: "/app/$workspaceId" });
      } else {
        void navigate({ to: "/app" });
      }
    },
  });
};
