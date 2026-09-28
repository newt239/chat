import { IconChartBar } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState";
import { PageHeader } from "#/features/layout/components/PageHeader";

// インサイトの中身は #16 で実装する
export const InsightsPage = () => {
  const { t } = useTranslation();
  return (
    <>
      <PageHeader icon={<IconChartBar />} title={t("shell.nav.insights")} />
      <EmptyState
        icon={<IconChartBar />}
        title={t("shell.nav.insights")}
        description={t("shell.comingSoon")}
      />
    </>
  );
};
