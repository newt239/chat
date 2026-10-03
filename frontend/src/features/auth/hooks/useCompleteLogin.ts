import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";

import { startSession } from "#/lib/session";
import { lastWorkspaceIdAtom } from "#/providers/store/workspace";

/** ログイン方法によらず、セッションを始めてアプリへ移動する。参加リンクから来たときはそのワークスペースを開く */
export const useCompleteLogin = (joinedWorkspaceId: string | null) => {
  const lastWorkspaceId = useAtomValue(lastWorkspaceIdAtom);
  const navigate = useNavigate();

  return (response: Parameters<typeof startSession>[0]) => {
    startSession(response);

    const workspaceId = joinedWorkspaceId ?? lastWorkspaceId;
    if (workspaceId) {
      void navigate({ params: { workspaceId }, to: "/app/$workspaceId" });
    } else {
      void navigate({ to: "/app" });
    }
  };
};
