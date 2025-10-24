import { useEffect, useMemo } from "react";

import { Avatar, Badge, Card, Loader, Stack, Text } from "@mantine/core";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { useChannels } from "@/features/channel/hooks/useChannel";
import {
  useWorkspaceSearchIndex,
  type WorkspaceSearchIndex,
} from "@/features/search/hooks/useWorkspaceSearchIndex";
import { MemberPanel } from "@/features/workspace/components/MemberPanel";
import { useMembers } from "@/features/workspace/hooks/useMembers";
import { store } from "@/lib/store";
import { isAuthenticatedAtom } from "@/lib/store/auth";
import { rightSidebarViewAtom, type RightSidebarView } from "@/lib/store/ui";
import { currentChannelIdAtom, setCurrentWorkspaceAtom } from "@/lib/store/workspace";

const SIDEBAR_CONTAINER_CLASS = "border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto";

const WorkspaceComponent = () => {
  const { workspaceId } = Route.useParams();
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);
  const rightSidebarView = useAtomValue(rightSidebarViewAtom);

  useEffect(() => {
    setCurrentWorkspace(workspaceId);
  }, [workspaceId, setCurrentWorkspace]);

  const layoutClassName =
    rightSidebarView.type === "hidden"
      ? "grid h-full min-h-0 gap-6 lg:grid-cols-[320px_1fr]"
      : "grid h-full min-h-0 gap-6 lg:grid-cols-[320px_1fr_280px]";

  return (
    <div className={layoutClassName}>
      <div className="space-y-6">
        <ChannelList workspaceId={workspaceId} />
      </div>
      <div className="min-h-0 w-full">
        <Outlet />
      </div>
      {rightSidebarView.type !== "hidden" && (
        <WorkspaceRightSidebar workspaceId={workspaceId} view={rightSidebarView} />
      )}
    </div>
  );
};

type WorkspaceRightSidebarProps = {
  workspaceId: string;
  view: RightSidebarView;
}

export const WorkspaceRightSidebar = ({ workspaceId, view }: WorkspaceRightSidebarProps) => {
  switch (view.type) {
    case "members":
      return <MemberPanel workspaceId={workspaceId} />;
    case "channel-info":
      return <ChannelInfoPanel workspaceId={workspaceId} channelId={view.channelId} />;
    case "thread":
      return <ThreadPanel threadId={view.threadId} />;
    case "user-profile":
      return <UserProfilePanel workspaceId={workspaceId} userId={view.userId} />;
    case "search":
      return (
        <SearchResultsPanel
          workspaceId={workspaceId}
          query={view.query}
          filter={view.filter}
        />
      );
    case "hidden":
      return null;
  }
};

type ChannelInfoPanelProps = {
  workspaceId: string;
  channelId?: string | null;
}

const ChannelInfoPanel = ({ workspaceId, channelId }: ChannelInfoPanelProps) => {
  const { data: channels, isLoading, isError, error } = useChannels(workspaceId);
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const effectiveChannelId = channelId ?? currentChannelId;
  const activeChannel = useMemo(() => {
    if (channels === undefined || effectiveChannelId === null) {
      return null;
    }
    return channels.find((candidate) => candidate.id === effectiveChannelId) ?? null;
  }, [channels, effectiveChannelId]);

  if (isLoading) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <div className="flex h-full items-center justify-center">
          <Loader size="sm" />
        </div>
      </div>
    );
  }

  if (isError) {
    const message = error instanceof Error ? error.message : "チャンネル情報の取得に失敗しました";
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  if (!activeChannel) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          チャンネル情報が見つかりませんでした
        </Text>
      </div>
    );
  }

  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <div>
          <Text size="sm" fw={600} className="mb-1">
            チャンネル情報
          </Text>
          <Text size="xs" c="dimmed">
            #{activeChannel.name}
          </Text>
        </div>
        {typeof activeChannel.description === "string" && activeChannel.description.length > 0 ? (
          <Text size="sm">{activeChannel.description}</Text>
        ) : (
          <Text size="sm" c="dimmed">
            説明は設定されていません
          </Text>
        )}
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            ステータス
          </Text>
          <Badge size="sm" variant="light" color={activeChannel.isPrivate ? "gray" : "blue"}>
            {activeChannel.isPrivate ? "プライベート" : "パブリック"}
          </Badge>
        </Stack>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            チャンネルID
          </Text>
          <Text size="xs" c="dimmed">
            {activeChannel.id}
          </Text>
        </Stack>
      </Stack>
    </div>
  );
};

type ThreadPanelProps = {
  threadId: string;
}

const ThreadPanel = ({ threadId }: ThreadPanelProps) => {
  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <div>
          <Text size="sm" fw={600}>
            スレッド
          </Text>
          <Text size="xs" c="dimmed">
            ID: {threadId}
          </Text>
        </div>
        <Text size="sm" c="dimmed">
          スレッドの詳細表示は現在準備中です。
        </Text>
      </Stack>
    </div>
  );
};

type UserProfilePanelProps = {
  workspaceId: string;
  userId: string;
}

const UserProfilePanel = ({ workspaceId, userId }: UserProfilePanelProps) => {
  const { data: members, isLoading, isError, error } = useMembers(workspaceId);
  const member = useMemo(() => {
    if (members === undefined) {
      return null;
    }
    return members.find((candidate) => candidate.userId === userId) ?? null;
  }, [members, userId]);

  if (isLoading) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <div className="flex h-full items-center justify-center">
          <Loader size="sm" />
        </div>
      </div>
    );
  }

  if (isError) {
    const message = error instanceof Error ? error.message : "ユーザープロフィールの取得に失敗しました";
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  if (member === null) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          指定されたユーザーが見つかりませんでした
        </Text>
      </div>
    );
  }

  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <div className="flex items-center gap-3">
          <Avatar src={member.avatarUrl ?? undefined} radius="xl" size="lg">
            {member.displayName.substring(0, 2).toUpperCase()}
          </Avatar>
          <div>
            <Text size="sm" fw={600}>
              {member.displayName}
            </Text>
            <Text size="xs" c="dimmed">
              {member.email}
            </Text>
          </div>
        </div>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            ロール
          </Text>
          <Badge size="sm" variant="light" color="gray">
            {member.role}
          </Badge>
        </Stack>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            ユーザーID
          </Text>
          <Text size="xs" c="dimmed">
            {member.userId}
          </Text>
        </Stack>
      </Stack>
    </div>
  );
};

type SearchFilter = "all" | "messages" | "channels" | "users";

type SearchResultsPanelProps = {
  workspaceId: string;
  query: string;
  filter: SearchFilter;
}

const SearchResultsPanel = ({ workspaceId, query, filter }: SearchResultsPanelProps) => {
  const trimmedQuery = query.trim();
  const lowercaseQuery = trimmedQuery.toLowerCase();
  const { data, isLoading, isError, error } = useWorkspaceSearchIndex(workspaceId);

  const filteredResults = useMemo<WorkspaceSearchIndex>(() => {
    const emptyResults: WorkspaceSearchIndex = { channels: [], members: [], messages: [] };

    if (data === undefined || lowercaseQuery.length === 0) {
      return emptyResults;
    }

    const includesQuery = (value: string | null | undefined) =>
      typeof value === "string" && value.toLowerCase().includes(lowercaseQuery);

    const channels = data.channels.filter(
      (channel) => includesQuery(channel.name) || includesQuery(channel.description)
    );

    const members = data.members.filter(
      (member) => includesQuery(member.displayName) || includesQuery(member.email ?? null)
    );

    const messages = data.messages.filter((message) => includesQuery(message.body));

    return { channels, members, messages };
  }, [data, lowercaseQuery]);

  if (trimmedQuery.length === 0) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          キーワードを入力すると検索結果が表示されます
        </Text>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <div className="flex h-full items-center justify-center">
          <Loader size="sm" />
        </div>
      </div>
    );
  }

  if (isError || data === undefined) {
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
    (shouldShowChannels && filteredResults.channels.length > 0) ||
    (shouldShowMembers && filteredResults.members.length > 0) ||
    (shouldShowMessages && filteredResults.messages.length > 0);

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
        {shouldShowChannels && filteredResults.channels.length > 0 && (
          <Stack gap="xs">
            <Text size="xs" c="dimmed">
              チャンネル
            </Text>
            {filteredResults.channels.map((channel) => (
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
        {shouldShowMembers && filteredResults.members.length > 0 && (
          <Stack gap="xs">
            <Text size="xs" c="dimmed">
              ユーザー
            </Text>
            {filteredResults.members.map((member) => (
              <Card key={member.userId} withBorder padding="md" radius="md">
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
        {shouldShowMessages && filteredResults.messages.length > 0 && (
          <Stack gap="xs">
            <Text size="xs" c="dimmed">
              メッセージ
            </Text>
            {filteredResults.messages.map((message) => (
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

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: () => {
    const isAuthenticated = store.get(isAuthenticatedAtom);
    if (!isAuthenticated) {
      throw redirect({ to: "/login" });
    }
  },
  component: WorkspaceComponent,
});
