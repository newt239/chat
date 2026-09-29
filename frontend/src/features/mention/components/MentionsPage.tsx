import { IconAt } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { MentionList } from "#/features/mention/components/MentionList";

export const MentionsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  return (
    <>
      <PageHeader icon={<IconAt />} title={t("shell.nav.mentions")}>
        <span className="hidden text-caption text-muted md:inline">{t("inbox.replyHint")}</span>
      </PageHeader>
      <MentionList workspaceId={workspaceId} />
    </>
  );
};
