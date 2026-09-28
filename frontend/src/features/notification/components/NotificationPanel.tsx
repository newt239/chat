import { formatRelativeTime } from "@chat/i18n";
import { IconAt, IconBell, IconMessage, IconMoodSmile, IconX } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Badge } from "#/components/ui/Badge";
import { IconButton } from "#/components/ui/IconButton";
import { cn, focusRing } from "#/components/ui/styles";
import {
  markNotificationAsReadAtom,
  notificationItemsAtom,
  removeNotificationAtom,
} from "#/providers/store/notification";
import { preferencesAtom } from "#/providers/store/preferences";

import type { NotificationItem } from "#/providers/store/notification";

const icons = {
  mention: IconAt,
  message: IconMessage,
  reaction: IconMoodSmile,
};

export const NotificationPanel = () => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const notifications = useAtomValue(notificationItemsAtom);
  const markAsRead = useSetAtom(markNotificationAsReadAtom);
  const removeNotification = useSetAtom(removeNotificationAtom);
  const navigate = useNavigate();

  const open = (notification: NotificationItem) => {
    if (!notification.isRead) {
      markAsRead(notification.id);
    }
    void navigate({
      params: { channelId: notification.channelId, workspaceId: notification.workspaceId },
      search: { message: notification.messageId },
      to: "/app/$workspaceId/$channelId",
    });
  };

  if (notifications.length === 0) {
    return (
      <div className="flex flex-col items-center gap-3 p-8 font-sans text-body text-muted">
        <IconBell aria-hidden className="size-10 text-subtle" />
        {t("notification.empty")}
      </div>
    );
  }

  const now = new Date();

  return (
    <ul className="m-0 flex list-none flex-col gap-0.5 overflow-y-auto p-1.5 font-sans">
      {notifications.map((notification) => {
        const TypeIcon = icons[notification.type];
        return (
          <li key={notification.id} className="flex items-start gap-1">
            <Button
              onPress={() => {
                open(notification);
              }}
              className={cn(
                "flex min-w-0 flex-1 cursor-pointer items-start gap-2.5 rounded-lg px-2 py-1.5 text-left text-text data-hovered:bg-hover",
                focusRing,
              )}
            >
              <span
                className={cn(
                  "mt-0.5 grid size-7 shrink-0 place-items-center rounded-md [&_svg]:size-4",
                  notification.isRead ? "bg-sunken text-muted" : "bg-accent-soft text-accent-text",
                )}
              >
                <TypeIcon aria-hidden />
              </span>
              <span className="flex min-w-0 flex-1 flex-col gap-0.5 leading-[1.35]">
                <span className="flex items-center gap-2">
                  <span
                    className={cn(
                      "truncate text-[13.5px]",
                      notification.isRead ? "font-normal" : "font-semibold",
                    )}
                  >
                    {notification.title}
                  </span>
                  <Badge tone={notification.isRead ? "tag" : "accent"}>
                    {t(`notification.type.${notification.type}`)}
                  </Badge>
                </span>
                <span className="truncate text-[12.5px] text-muted">{notification.message}</span>
                <span className="flex justify-between gap-2 text-[11.5px] text-subtle">
                  <span className="truncate">
                    #{notification.channelName}
                    {notification.userName && ` · ${notification.userName}`}
                  </span>
                  <span className="shrink-0">
                    {formatRelativeTime(new Date(notification.timestamp), now, locale)}
                  </span>
                </span>
              </span>
            </Button>
            <IconButton
              label={t("notification.remove")}
              className="mt-1 size-7 [&_svg]:size-3.5"
              onPress={() => {
                removeNotification(notification.id);
              }}
            >
              <IconX />
            </IconButton>
          </li>
        );
      })}
    </ul>
  );
};
