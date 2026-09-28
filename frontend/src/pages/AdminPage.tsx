import { IconShieldCheck } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState";
import { PageHeader } from "#/features/layout/components/PageHeader";
import { useIsWorkspaceAdmin } from "#/features/workspace/hooks/useIsWorkspaceAdmin";

// 管理画面の中身は #16 で実装する
export const AdminPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const isAdmin = useIsWorkspaceAdmin(workspaceId);
  return (
    <>
      <PageHeader icon={<IconShieldCheck />} title={t("shell.nav.admin")} />
      <EmptyState
        icon={<IconShieldCheck />}
        title={t("shell.nav.admin")}
        description={isAdmin ? t("shell.comingSoon") : t("shell.admin.forbidden")}
      />
    </>
  );
};
