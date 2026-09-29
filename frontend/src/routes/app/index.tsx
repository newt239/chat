import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { WorkspaceSelection } from "#/features/workspace/components/WorkspaceSelection";

export const Route = createFileRoute("/app/")({
  component: WorkspaceSelection,
  validateSearch: z.object({
    dialog: z.enum(["create-workspace"]).optional().catch(undefined),
    // ホーム画面のアイコンのショートカット（manifest の shortcuts）から開く画面
    open: z.enum(["dms", "activity", "search"]).optional().catch(undefined),
  }),
});
