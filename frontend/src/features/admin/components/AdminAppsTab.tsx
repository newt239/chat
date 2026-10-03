import { IconPlus } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { AppRow } from "#/features/app/components/AppRow";
import { useApps } from "#/features/app/hooks/useApps";
import { openDialog } from "#/lib/overlaySearch";

type AdminAppsTabProps = {
  workspaceId: string;
};

// ワークスペースのアプリ一覧。管理者はすべてのアプリを編集・削除できる
export const AdminAppsTab = ({ workspaceId }: AdminAppsTabProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data } = useApps(workspaceId);
  const apps = data?.apps ?? [];

  return (
    <section className="flex flex-col gap-2">
      <div className="flex items-center justify-between gap-2">
        <p className="m-0 text-caption text-muted">{t("app.hint")}</p>
        <Button
          size="sm"
          onPress={() => {
            void navigate({ search: openDialog({ dialog: "add-app" }), to: "." });
          }}
        >
          <IconPlus aria-hidden />
          {t("app.create")}
        </Button>
      </div>
      {apps.length === 0 ? (
        <p className="m-0 text-caption text-muted">{t("app.empty")}</p>
      ) : (
        <ul className="m-0 flex list-none flex-col p-0">
          {apps.map((app) => (
            <AppRow key={app.id} app={app} actions={null} />
          ))}
        </ul>
      )}
    </section>
  );
};
