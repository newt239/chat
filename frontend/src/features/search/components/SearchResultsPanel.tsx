import { Avatar, Badge, Card, Loader, Stack, Text } from "@mantine/core";
import { useSetAtom } from "jotai";

import { useWorkspaceSearch } from "@/features/search/hooks/useWorkspaceSearchIndex";
import { setRightSidePanelViewAtom } from "@/providers/store/ui";

const SIDEBAR_CONTAINER_CLASS = "border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto";

type SearchFilter = "all" | "messages" | "channels" | "users";

type SearchResultsPanelProps = {
  workspaceId: string;
  query: string;
  filter: SearchFilter;
};

export const SearchResultsPanel = ({ workspaceId, query, filter }: SearchResultsPanelProps) => {
  const trimmedQuery = query.trim();
  const { data, isLoading, isFetching, error } = useWorkspaceSearch({
    workspaceId,
    query,
    filter,
    page: 1,
    perPage: 10,
  });
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  const handleUserClick = (userId: string) => {
    setRightSidePanelView({ type: "user-profile", userId });
  };

  if (trimmedQuery.length === 0) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          キーワードを入力すると検索結果が表示されます
        </Text>
      </div>
    );
  }

  if (isLoading || isFetching) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <div className="flex h-full items-center justify-center">
          <Loader size="sm" />
        </div>
      </div>
    );
  }

  if (error || data === undefined) {
    const message = error instanceof Error ? error.message : "検索結果の取得に失敗しました";
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  const shouldShowChannels = filter === "all" || filter === "channels";
  const shouldShowMembers = filter === "all" || filter === "users";
  const shouldShowMessages = filter === "all" || filter === "messages";

  const hasResults =
    (shouldShowChannels && data.channels.items.length > 0) ||
    (shouldShowMembers && data.users.items.length > 0) ||
    (shouldShowMessages && data.messages.items.length > 0);

  if (!hasResults) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          「{trimmedQuery}」に一致する結果は見つかりませんでした
        </Text>
      </div>
    );
  }

  const dateTimeFormatter = new Intl.DateTimeFormat("ja-JP", {
    dateStyle: "short",
    timeStyle: "short",
  });

  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <Text size="sm" fw={600}>
          検索結果
        </Text>
        {shouldShowChannels && data.channels.items.length > 0 && (
          <Stack gap="xs">
            <Text size="xs" c="dimmed">
              チャンネル
            </Text>
            {data.channels.items.map((channel) => (
              <Card key={channel.id} withBorder padding="md" radius="md">
                <Stack gap="4">
                  <Text size="sm" fw={600}>
                    #{channel.name}
                  </Text>
                  {typeof channel.description === "string" && channel.description.length > 0 ? (
                    <Text size="xs" c="dimmed">
                      {channel.description}
                    </Text>
                  ) : null}
                  <Badge size="xs" variant="light" color={channel.isPrivate ? "gray" : "blue"}>
                    {channel.isPrivate ? "プライベート" : "パブリック"}
                  </Badge>
                </Stack>
              </Card>
            ))}
          </Stack>
        )}
        {shouldShowMembers && data.users.items.length > 0 && (
          <Stack gap="xs">
            <Text size="xs" c="dimmed">
              ユーザー
            </Text>
            {data.users.items.map((member) => (
              <Card
                key={member.userId}
                withBorder
                padding="md"
                radius="md"
                className="cursor-pointer hover:bg-gray-100 transition-colors"
                onClick={() => handleUserClick(member.userId)}
              >
                <div className="flex items-center gap-3">
                  <Avatar src={member.avatarUrl ?? undefined} radius="xl" size="md" />
                  <div className="flex-1">
                    <Text size="sm" fw={600}>
                      {member.displayName}
                    </Text>
                    <Text size="xs" c="dimmed">
                      {member.email}
                    </Text>
                  </div>
                </div>
              </Card>
            ))}
          </Stack>
        )}
        {shouldShowMessages && data.messages.items.length > 0 && (
          <Stack gap="xs">
            <Text size="xs" c="dimmed">
              メッセージ
            </Text>
            {data.messages.items.map((message) => (
              <Card key={message.id} withBorder padding="md" radius="md">
                <Stack gap="4">
                  <Text size="xs" c="dimmed">
                    {dateTimeFormatter.format(new Date(message.createdAt))}
                  </Text>
                  <Text size="sm">{message.body}</Text>
                </Stack>
              </Card>
            ))}
          </Stack>
        )}
      </Stack>
    </div>
  );
};
