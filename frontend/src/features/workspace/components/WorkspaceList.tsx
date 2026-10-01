import { getRouteApi, useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton/LinkButton";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";

import { useWorkspaces } from "../hooks/useWorkspace";
import { CreateWorkspaceModal } from "./CreateWorkspaceModal";
import { PublicWorkspaceList } from "./PublicWorkspaceList";
import { WorkspaceLogo } from "./WorkspaceLogo";

const route = getRouteApi("/app/");

export const WorkspaceList = () => {
  const { t } = useTranslation();
  const { data: workspaces, isLoading, isError } = useWorkspaces();
  const navigate = useNavigate();
  const { dialog } = route.useSearch();

  return (
    <div className="flex w-full max-w-3xl flex-col gap-6">
      <section className="flex flex-col gap-3">
        <header className="flex items-center justify-between gap-2">
          <h1 className="m-0 text-title">{t("workspace.list.title")}</h1>
          <LinkButton to="/app" search={{ dialog: "create-workspace" }}>
            {t("workspace.create.title")}
          </LinkButton>
        </header>
        {isLoading && <Skeleton className="h-16 w-full" />}
        {isError && <p className="m-0 text-danger">{t("workspace.list.loadFailed")}</p>}
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          {workspaces?.map((workspace) => (
            <article
              key={workspace.id}
              className="flex items-center gap-3 rounded-lg border border-border bg-surface p-4"
            >
              <WorkspaceLogo name={workspace.name} iconUrl={workspace.iconUrl} />
              <div className="flex min-w-0 flex-1 flex-col">
                <b className="truncate text-body-strong">{workspace.name}</b>
                {workspace.description && (
                  <span className="truncate text-caption text-muted">{workspace.description}</span>
                )}
              </div>
              <LinkButton
                variant="secondary"
                to="/app/$workspaceId"
                params={{ workspaceId: workspace.id }}
              >
                {t("workspace.list.open")}
              </LinkButton>
            </article>
          ))}
        </div>
      </section>

      <PublicWorkspaceList />

      <CreateWorkspaceModal
        isOpen={dialog === "create-workspace"}
        onOpenChange={(isOpen) => {
          void navigate({
            search: { dialog: isOpen ? "create-workspace" : undefined },
            to: "/app",
          });
        }}
      />
    </div>
  );
};
