import type { ReactNode } from "react";

import { formatRelativeTime } from "@chat/i18n/format";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Badge } from "#/components/ui/Badge/Badge";
import { usePreferences } from "#/hooks/usePreferences";
import { toDate } from "#/lib/timestamp";

import type { App } from "#/gen/chat/v1/app_service_pb";

type AppRowProps = {
  app: App;
  // 行の右端に置く操作
  actions: ReactNode;
};

// アプリの名前・作成者・最終使用日時を 1 行で出す
export const AppRow = ({ app, actions }: AppRowProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();

  return (
    <li className="flex items-center gap-2 rounded-md px-2 py-1">
      <Avatar name={app.name} src={app.avatarUrl} size={28} />
      <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
        <span className="flex items-center gap-1.5 truncate text-[13.5px]">
          {app.name}
          {app.isOfficial && <Badge tone="tag">{t("app.official")}</Badge>}
        </span>
        <small className="truncate text-[11.5px] text-subtle">
          {app.isOfficial
            ? app.description
            : [
                t("app.createdBy", { name: app.createdBy?.displayName ?? "" }),
                app.lastUsedAt
                  ? t("app.lastUsed", {
                      time: formatRelativeTime(toDate(app.lastUsedAt), new Date(), locale),
                    })
                  : t("app.neverUsed"),
              ].join(" · ")}
        </small>
      </span>
      <span className="flex shrink-0 [&_button]:size-7 [&_svg]:size-4! max-md:gap-1 max-md:[&_button]:size-11">
        {actions}
      </span>
    </li>
  );
};
