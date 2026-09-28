import { IconBell } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/features/layout/components/PageHeader";
import { MentionList } from "#/features/mention/components/MentionList";

// モバイルの「通知」タブ。中身はメンション一覧
export const ActivityPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  return (
    <>
      <PageHeader icon={<IconBell />} title={t("shell.nav.activity")} />
      <MentionList workspaceId={workspaceId} />
    </>
  );
};
