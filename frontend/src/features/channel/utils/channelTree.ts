import type { Channel } from "#/gen/chat/v1/channel_service_pb";

export type ChannelTreeNode = {
  channel: Channel;
  children: ChannelTreeNode[];
};

// parent_id でつなぐ。親が一覧にない（閲覧できない）場合は最上位に置く
export const buildChannelTree = (channels: readonly Channel[]) => {
  const ids = new Set(channels.map((channel) => channel.id));
  const childrenOf = Map.groupBy(channels, (channel) =>
    channel.parentId !== undefined && ids.has(channel.parentId) ? channel.parentId : "",
  );
  const build = (parentId: string): ChannelTreeNode[] =>
    (childrenOf.get(parentId) ?? [])
      .toSorted((a, b) => a.name.localeCompare(b.name))
      .map((channel) => ({ channel, children: build(channel.id) }));
  return build("");
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
