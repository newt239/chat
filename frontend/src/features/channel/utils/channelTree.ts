import { toDate } from "#/lib/timestamp";

import type { Channel } from "#/gen/chat/v1/channel_service_pb";
import type { Preferences } from "#/providers/store/preferences";

export type ChannelTreeNode = {
  channel: Channel;
  children: ChannelTreeNode[];
};

const activityOf = (channel: Channel) =>
  toDate(channel.lastMessageAt ?? channel.createdAt).getTime();

// ミュート中はカテゴリ内の最後に回し、その中で名前順か新しいメッセージ順に並べる
const compareChannels = (order: Preferences["channelSortOrder"]) => (a: Channel, b: Channel) =>
  Number(a.isMuted) - Number(b.isMuted) ||
  (order === "recentActivity" ? activityOf(b) - activityOf(a) : 0) ||
  a.name.localeCompare(b.name);

// parent_id でつなぐ。親が一覧にない（閲覧できない・別のカテゴリにある）場合は最上位に置く
export const buildChannelTree = (channels: readonly Channel[]) => {
  const ids = new Set(channels.map((channel) => channel.id));
  const childrenOf = Map.groupBy(channels, (channel) =>
    channel.parentId !== undefined && ids.has(channel.parentId) ? channel.parentId : "",
  );
  const build = (parentId: string): ChannelTreeNode[] =>
    (childrenOf.get(parentId) ?? [])
      .toSorted(compareChannels("default"))
      .map((channel) => ({ channel, children: build(channel.id) }));
  return build("");
};

// 新しいメッセージ順では階層を無視して、参加中のチャンネルだけを並べる
export const sortChannelsByActivity = (channels: readonly Channel[]) =>
  channels.filter((channel) => channel.isMember).toSorted(compareChannels("recentActivity"));

// 自分への割り当て、なければ最も近い祖先の割り当てに従う。どれもなければ既定のカテゴリ（null）
export const categoryOfChannel = (
  channel: Channel,
  channels: readonly Channel[],
  categoryByChannel: ReadonlyMap<string, string>,
) => {
  const byId = new Map(channels.map((candidate) => [candidate.id, candidate]));
  let current: Channel | undefined = channel;
  while (current !== undefined) {
    const categoryId = categoryByChannel.get(current.id);
    if (categoryId !== undefined) {
      return categoryId;
    }
    current = current.parentId === undefined ? undefined : byId.get(current.parentId);
  }
  return null;
};

// 自分と子孫の未読の合計。ミュート中のチャンネルは数えない
export const sumUnread = (node: ChannelTreeNode) => {
  let unreadCount = 0;
  let hasMention = false;
  const walk = ({ channel, children }: ChannelTreeNode) => {
    if (!channel.isMuted) {
      unreadCount += channel.unreadCount;
      hasMention ||= channel.hasMention;
    }
    for (const child of children) {
      walk(child);
    }
  };
  walk(node);
  return { hasMention, unreadCount };
};

export const isDescendantPath = (ancestor: string, path: string) => path.startsWith(`${ancestor}/`);

// 表示中の親から見た相対パス（dev から見た dev/frontend/web は frontend/web）
export const relativePath = (ancestor: string, path: string) =>
  isDescendantPath(ancestor, path) ? path.slice(ancestor.length + 1) : path;
