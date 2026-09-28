import { IconChartBar } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { InsightsDashboard } from "#/features/insights/components/InsightsDashboard";
import { PageHeader } from "#/features/layout/components/PageHeader";

export const InsightsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  return (
    <section className="flex h-full min-h-0 flex-col bg-surface font-sans text-text">
      <PageHeader icon={<IconChartBar />} title={t("insights.title")} />
      <div className="min-h-0 flex-1 overflow-y-auto">
        <InsightsDashboard workspaceId={workspaceId} />
      </div>
    </section>
  );
};
