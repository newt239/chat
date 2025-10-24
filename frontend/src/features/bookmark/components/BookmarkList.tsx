import { useNavigate } from "@tanstack/react-router";

import { Button, Text, Stack, ScrollArea, Group, Avatar } from "@mantine/core";
import { IconBookmark, IconMessage } from "@tabler/icons-react";

import { useBookmarks } from "../hooks/useBookmarks";

export const BookmarkList = () => {
  const { data: bookmarks, isLoading, error } = useBookmarks();
  const navigate = useNavigate();

  const handleBookmarkClick = (channelId: string, messageId: string) => {
    navigate({
      to: "/app/$workspaceId/channels/$channelId",
      params: { workspaceId: "current", channelId },
      search: { messageId },
    });
  };

  if (isLoading) {
    return (
      <div className="p-4">
        <Text size="sm" c="dimmed">
          読み込み中...
        </Text>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <Text size="sm" c="red">
          エラーが発生しました
        </Text>
      </div>
    );
  }

  if (!bookmarks?.bookmarks || bookmarks.bookmarks.length === 0) {
    return (
      <div className="p-4 text-center">
        <IconBookmark size={48} className="mx-auto mb-4 text-gray-400" />
        <Text size="sm" c="dimmed">
          ブックマークされたメッセージはありません
        </Text>
      </div>
    );
  }

  return (
    <div className="p-4">
      <div className="mb-4">
        <Text size="lg" fw={600}>
          ブックマーク
        </Text>
        <Text size="sm" c="dimmed">
          {bookmarks.bookmarks.length}件のメッセージ
        </Text>
      </div>

      <ScrollArea h={400}>
        <Stack gap="xs">
          {bookmarks.bookmarks.map((bookmark) => (
            <Button
              key={`${bookmark.userId}-${bookmark.messageId}`}
              variant="subtle"
              className="h-auto p-3 text-left justify-start"
              onClick={() => handleBookmarkClick(bookmark.message.channelId, bookmark.message.id)}
            >
              <Group className="w-full" gap="sm" align="flex-start">
                <Avatar size="sm" color="blue">
                  <IconMessage size={16} />
                </Avatar>
                <div className="flex-1 min-w-0">
                  <Text size="sm" fw={500} lineClamp={2}>
                    {bookmark.message.body}
                  </Text>
                  <Text size="xs" c="dimmed" mt={4}>
                    {new Date(bookmark.createdAt).toLocaleDateString("ja-JP", {
                      year: "numeric",
                      month: "short",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </Text>
                </div>
              </Group>
            </Button>
          ))}
        </Stack>
      </ScrollArea>
    </div>
  );
};
