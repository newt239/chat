import { useChannels } from "#/features/channel/hooks/useChannel";
import { useMembers } from "#/features/member/hooks/useMembers";
import { parseSearchQuery } from "#/features/search/utils/searchQuery";

import type { WorkspaceMember } from "#/gen/chat/v1/workspace_service_pb";

const same = (a: string, b: string | undefined) =>
  b !== undefined && a.toLowerCase() === b.toLowerCase();

const matchesMember = (name: string, member: WorkspaceMember) =>
  same(name, member.displayName) ||
  same(name, member.nickname) ||
  same(name, member.email.split("@")[0]);

// 修飾子の名前（from:@名前 / in:#チャンネル名）をメンバー・チャンネルに解決する
export const useResolvedSearchQuery = (workspaceId: string, raw: string) => {
  const { data: members } = useMembers(workspaceId);
  const { data: channels } = useChannels(workspaceId);
  const query = parseSearchQuery(raw);

  const users = query.from.map((name) => ({
    member: members?.find((member) => matchesMember(name, member)),
    name,
  }));
  const inChannels = query.in.map((name) => ({
    channel: channels?.find((channel) => same(name, channel.name)),
    name,
  }));
  const hasDescendants = inChannels.some(({ channel }) =>
    channels?.some((other) => channel !== undefined && other.name.startsWith(`${channel.name}/`)),
  );

  return {
    channels: channels ?? [],
    hasDescendants,
    inChannels,
    isResolving: members === undefined || channels === undefined,
    members: members ?? [],
    query,
    users,
  };
};
