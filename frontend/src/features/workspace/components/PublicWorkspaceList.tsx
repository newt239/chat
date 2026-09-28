import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import {
  useJoinPublicWorkspace,
  usePublicWorkspaces,
} from "#/features/workspace/hooks/usePublicWorkspaces";

import { WorkspaceLogo } from "./WorkspaceLogo";

export const PublicWorkspaceList = () => {
  const { t } = useTranslation();
  const { data: workspaces } = usePublicWorkspaces();
  const join = useJoinPublicWorkspace();
  const joinable = workspaces?.filter((workspace) => !workspace.isJoined) ?? [];

  if (joinable.length === 0) {
    return null;
  }

  return (
    <section className="flex flex-col gap-3">
      <h2 className="m-0 text-body-strong">{t("workspace.list.public")}</h2>
      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        {joinable.map((workspace) => (
          <article
            key={workspace.id}
            className="flex items-center gap-3 rounded-lg border border-border bg-surface p-4"
          >
            <WorkspaceLogo name={workspace.name} />
            <div className="flex min-w-0 flex-1 flex-col">
              <b className="truncate text-body-strong">{workspace.name}</b>
              <span className="truncate text-caption text-muted">
                {t("workspace.list.memberCount", { count: workspace.memberCount })}
                {workspace.description && ` · ${workspace.description}`}
              </span>
            </div>
            <Button
              variant="secondary"
              isPending={join.isPending && join.variables.workspaceId === workspace.id}
              onPress={() => {
                join.mutate({ workspaceId: workspace.id });
              }}
            >
              {t("workspace.list.join")}
            </Button>
          </article>
        ))}
      </div>
    </section>
  );
};
