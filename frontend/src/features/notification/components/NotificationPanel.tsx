import { useCallback } from "react";

import { ActionIcon, Badge, Card, ScrollArea, Stack, Text } from "@mantine/core";
import { IconBell, IconX } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import {
  markNotificationAsReadAtom,
  removeNotificationAtom,
  type NotificationItem,
  notificationItemsAtom,
} from "@/providers/store/notification";

export const NotificationPanel = () => {
  const notifications = useAtomValue(notificationItemsAtom);
  const markAsRead = useSetAtom(markNotificationAsReadAtom);
  const removeNotification = useSetAtom(removeNotificationAtom);
  const navigate = useNavigate();

  const handleNotificationClick = useCallback(
    (notification: NotificationItem) => {
      // 既読にする
      if (!notification.isRead) {
        markAsRead(notification.id);
      }

      // チャンネルに遷移
      navigate({
        to: "/app/$workspaceId/$channelId",
        params: {
          workspaceId: notification.workspaceId,
          channelId: notification.channelId,
        },
        search: notification.messageId ? { message: notification.messageId } : {},
      });
    },
    [markAsRead, navigate]
  );

  const handleRemoveNotification = useCallback(
    (notificationId: string, event: React.MouseEvent) => {
      event.stopPropagation();
      removeNotification(notificationId);
    },
    [removeNotification]
  );

  const formatTimestamp = (timestamp: Date) => {
    const now = new Date();
    const diff = now.getTime() - timestamp.getTime();
    const minutes = Math.floor(diff / (1000 * 60));
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (minutes < 1) {
      return "たった今";
    } else if (minutes < 60) {
      return `${minutes}分前`;
    } else if (hours < 24) {
      return `${hours}時間前`;
    } else if (days < 7) {
      return `${days}日前`;
    } else {
      return timestamp.toLocaleDateString("ja-JP");
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case "mention":
        return <IconBell size={16} />;
      case "message":
        return <IconBell size={16} />;
      case "reaction":
        return <IconBell size={16} />;
      default:
        return <IconBell size={16} />;
    }
  };

  const getNotificationColor = (type: string, isRead: boolean) => {
    if (isRead) return "gray";
    switch (type) {
      case "mention":
        return "red";
      case "message":
        return "blue";
      case "reaction":
        return "green";
      default:
        return "gray";
    }
  };

  return (
    <div className="h-full flex flex-col">
      {/* 通知一覧 */}
      <ScrollArea className="flex-1">
        {notifications.length === 0 ? (
          <div className="p-8 text-center">
            <IconBell size={48} className="mx-auto text-gray-400 mb-4" />
            <Text c="dimmed" size="sm">
              通知はありません
            </Text>
          </div>
        ) : (
          <Stack gap="xs" className="p-2">
            {notifications.map((notification) => (
              <Card
                key={notification.id}
                className={`cursor-pointer transition-colors ${
                  notification.isRead ? "bg-gray-50" : "bg-white"
                } hover:bg-gray-100`}
                padding="sm"
                onClick={() => handleNotificationClick(notification)}
              >
                <div className="flex items-start gap-3">
                  <div
                    className={`p-1 rounded ${notification.isRead ? "bg-gray-200" : "bg-blue-100"}`}
                  >
                    {getNotificationIcon(notification.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <Text size="sm" fw={notification.isRead ? 400 : 600} className="truncate">
                        {notification.title}
                      </Text>
                      <Badge
                        size="xs"
                        color={getNotificationColor(notification.type, notification.isRead)}
                      >
                        {notification.type === "mention"
                          ? "メンション"
                          : notification.type === "message"
                            ? "メッセージ"
                            : "リアクション"}
                      </Badge>
                    </div>
                    <Text size="xs" c="dimmed" className="truncate mb-1">
                      {notification.message}
                    </Text>
                    <div className="flex items-center justify-between">
                      <Text size="xs" c="dimmed">
                        #{notification.channelName}
                        {notification.userName && ` • ${notification.userName}`}
                      </Text>
                      <Text size="xs" c="dimmed">
                        {formatTimestamp(notification.timestamp)}
                      </Text>
                    </div>
                  </div>
                  <ActionIcon
                    size="sm"
                    variant="subtle"
                    color="gray"
                    onClick={(e) => handleRemoveNotification(notification.id, e)}
                  >
                    <IconX size={14} />
                  </ActionIcon>
                </div>
              </Card>
            ))}
          </Stack>
        )}
      </ScrollArea>
    </div>
  );
};
