import { IconBell } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/features/layout/components/PageHeader";
import { NotificationPanel } from "#/features/notification/components/NotificationPanel";

// モバイルの「通知」タブ。WebSocket で受け取った新着とメンションを並べる
export const ActivityPage = () => {
  const { t } = useTranslation();
  return (
    <>
      <PageHeader icon={<IconBell />} title={t("shell.nav.activity")} />
      <div className="min-h-0 flex-1 overflow-y-auto">
        <NotificationPanel />
      </div>
    </>
  );
};
