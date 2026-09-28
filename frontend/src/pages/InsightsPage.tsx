import { IconChartBar } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { InsightsDashboard } from "#/features/insights/components/InsightsDashboard";

export const InsightsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  return (
    <section className="flex h-full min-h-0 flex-col bg-surface font-sans text-text">
      <header className="flex h-12 shrink-0 items-center gap-2 border-b border-border px-[18px] max-md:px-3">
        <IconChartBar aria-hidden className="size-4 shrink-0 text-muted" />
        <h1 className="m-0 shrink-0 text-[15px] font-bold">{t("insights.title")}</h1>
        <span className="min-w-0 truncate text-caption text-muted max-md:hidden">
          {t("insights.subtitle")}
        </span>
      </header>
      <div className="min-h-0 flex-1 overflow-y-auto">
        <InsightsDashboard workspaceId={workspaceId} />
      </div>
    </section>
  );
};
