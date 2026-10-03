import { getRouteApi, useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton/LinkButton";

import { CreateWorkspaceModal } from "./CreateWorkspaceModal";
import { PublicWorkspaceList } from "./PublicWorkspaceList";

const route = getRouteApi("/app/");

export const WorkspaceList = () => {
  const { t } = useTranslation();
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
      </section>

      <PublicWorkspaceList />

      {dialog === "create-workspace" && (
        <CreateWorkspaceModal
          onClose={() => {
            void navigate({ search: { dialog: undefined }, to: "/app" });
          }}
        />
      )}
    </div>
  );
};
