import { Text, Stack, ScrollArea, Card } from "@mantine/core";
import { IconBookmark } from "@tabler/icons-react";
import { Link } from "@tanstack/react-router";
import { useAtomValue } from "jotai";

import { toDate } from "#/lib/timestamp";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

import { useBookmarks } from "../hooks/useBookmarks";

export const BookmarkList = () => {
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const { data: bookmarks, isLoading, error } = useBookmarks();

  if (!workspaceId) {
    return null;
  }

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

  if (!bookmarks || bookmarks.length === 0) {
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
        {bookmarks.map((bookmark) => (
          <Card
            key={`${bookmark.userId}-${bookmark.message.id}`}
            withBorder
            padding="md"
            radius="md"
            renderRoot={(props) => (
              <Link
                {...props}
                to="/app/$workspaceId/$channelId"
                params={{ channelId: bookmark.message.channelId, workspaceId }}
                search={{ message: bookmark.message.id }}
              />
            )}
            className="h-auto text-left justify-start"
          >
            <div className="flex-1 min-w-0">
              <Text size="sm" fw={500} className="whitespace-pre-wrap">
                {bookmark.message.body}
              </Text>
              <Text size="xs" c="dimmed" mt={4}>
                {toDate(bookmark.createdAt).toLocaleDateString("ja-JP", {
                  day: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                  month: "short",
                  year: "numeric",
                })}
              </Text>
            </div>
          </Card>
        ))}
      </Stack>
    </ScrollArea>
  );
};
