import { IconFilePencil } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { Tab } from "#/components/ui/Tab/Tab";
import { TabList } from "#/components/ui/TabList/TabList";
import { TabPanel } from "#/components/ui/TabPanel/TabPanel";
import { Tabs } from "#/components/ui/Tabs/Tabs";
import { DraftList } from "#/features/draft/components/DraftList";
import { ScheduledMessageList } from "#/features/schedule/components/ScheduledMessageList";

const routeApi = getRouteApi("/app/$workspaceId/drafts");

// 下書き・予約済み・予約から送信済みのメッセージをタブで切り替えて見る
export const DraftsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = routeApi.useParams();

  return (
    <>
      <PageHeader icon={<IconFilePencil />} title={t("draft.page.title")} />
      <Tabs className="min-h-0 flex-1">
        <TabList aria-label={t("draft.page.title")}>
          <Tab id="drafts">{t("draft.page.tabs.drafts")}</Tab>
          <Tab id="scheduled">{t("draft.page.tabs.scheduled")}</Tab>
          <Tab id="sent">{t("draft.page.tabs.sent")}</Tab>
        </TabList>
        <TabPanel id="drafts" className="overflow-y-auto">
          <DraftList workspaceId={workspaceId} />
        </TabPanel>
        <TabPanel id="scheduled" className="overflow-y-auto">
          <ScheduledMessageList workspaceId={workspaceId} mode="pending" />
        </TabPanel>
        <TabPanel id="sent" className="overflow-y-auto">
          <ScheduledMessageList workspaceId={workspaceId} mode="sent" />
        </TabPanel>
      </Tabs>
    </>
  );
};
