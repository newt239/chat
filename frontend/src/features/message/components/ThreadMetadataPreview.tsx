import { Avatar, Group, Text } from "@mantine/core";
import { IconMessageCircle } from "@tabler/icons-react";

import type { ThreadMetadata } from "../types";

type ThreadMetadataPreviewProps = {
  metadata: ThreadMetadata;
  onClick: () => void;
};

export const ThreadMetadataPreview = ({
  metadata,
  onClick,
}: ThreadMetadataPreviewProps) => {
  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) {
      return "たった今";
    }

    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) {
      return `${diffInMinutes}分前`;
    }

    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) {
      return `${diffInHours}時間前`;
    }

    const diffInDays = Math.floor(diffInHours / 24);
    if (diffInDays < 7) {
      return `${diffInDays}日前`;
    }

    return date.toLocaleDateString("ja-JP");
  };

  return (
    <div
      className="ml-12 mt-1 cursor-pointer"
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onClick();
        }
      }}
    >
      <Group gap="xs" className="hover:opacity-70 transition-opacity">
        <IconMessageCircle size={16} className="text-blue-600" />
        <Text size="sm" c="blue">
          {metadata.replyCount}件の返信
        </Text>
        {metadata.lastReplyAt && (
          <Text size="sm" c="dimmed">
            最終返信: {formatRelativeTime(metadata.lastReplyAt)}
          </Text>
        )}
        {metadata.lastReplyUser && (
          <Avatar
            src={metadata.lastReplyUser.avatarUrl ?? undefined}
            alt={metadata.lastReplyUser.displayName}
            size="xs"
            radius="xl"
            color="blue"
          >
            {metadata.lastReplyUser.displayName.charAt(0).toUpperCase()}
          </Avatar>
        )}
      </Group>
    </div>
  );
};
