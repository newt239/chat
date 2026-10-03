import { createFileRoute } from "@tanstack/react-router";

import { JoinWorkspace } from "#/features/auth/components/JoinWorkspace";
import { ensureSession } from "#/lib/session";

// ログイン済みなら今のアカウントのまま参加させるため、Cookie でセッションを取り直しておく
export const Route = createFileRoute("/join/$workspaceId")({
  beforeLoad: async () => {
    await ensureSession();
  },
  component: JoinWorkspace,
});
