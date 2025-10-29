import { Text, Stack, ScrollArea, Card } from "@mantine/core";
import { IconBookmark } from "@tabler/icons-react";
import { Link } from "@tanstack/react-router";

import { useBookmarks } from "../hooks/useBookmarks";

export const BookmarkList = () => {
  const { data: bookmarks, isLoading, error } = useBookmarks();

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
    <ScrollArea h={400}>
      <Stack gap="xs" p="xs">
        {bookmarks?.bookmarks
          ?.filter((bookmark) => bookmark.message)
          .map((bookmark) => (
            <Card
              key={`${bookmark.userId}-${bookmark.message.id}`}
              withBorder
              padding="md"
              radius="md"
              component={Link}
              to={`/app/$workspaceId/channels/$channelId?messageId=${bookmark.message.id}`}
              className="h-auto text-left justify-start"
            >
              <div className="flex-1 min-w-0">
                <Text size="sm" fw={500} className="whitespace-pre-wrap">
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
            </Card>
          ))}
      </Stack>
    </ScrollArea>
  );
};
