import { IconAt } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState";
import { PageHeader } from "#/features/layout/components/PageHeader";

// メンション一覧の中身は #15 で実装する
export const MentionsPage = () => {
  const { t } = useTranslation();
  return (
    <>
      <PageHeader icon={<IconAt />} title={t("shell.nav.mentions")} />
      <EmptyState
        icon={<IconAt />}
        title={t("shell.mentions.emptyTitle")}
        description={t("shell.mentions.emptyDescription")}
      />
    </>
  );
};
