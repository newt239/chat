import type { ReactNode } from "react";

import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";

import { Link } from "#/components/ui/Link/Link";
import { cn } from "#/components/ui/styles/styles";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { chipClassName } from "./chipClassName";

type ChannelLinkProps = {
  "data-channel": string;
  children?: ReactNode;
};

export const ChannelLink = ({ "data-channel": channelName }: ChannelLinkProps) => {
  const { workspaceId } = useParams({ strict: false });
  const { data: channels } = useChannels(workspaceId ?? null);
  const listed = channels?.find((item) => item.name === channelName);
  // 未参加の公開チャンネルは一覧にないため、参加できるチャンネルから探す
  const { data: browsable } = useQuery(
    ChannelService.method.listBrowsableChannels,
    channels !== undefined && listed === undefined && workspaceId !== undefined
      ? { workspaceId }
      : skipToken,
    { select: (res) => res.channels.find((item) => item.channel?.name === channelName)?.channel },
  );
  const channel = listed ?? browsable;

  if (workspaceId === undefined || channel === undefined) {
    return <span className={cn(chipClassName, "cursor-default")}>#{channelName}</span>;
  }

  return (
    <Link
      to="/app/$workspaceId/$channelId"
      params={{ channelId: channel.id, workspaceId }}
      className={chipClassName}
    >
      #{channelName}
    </Link>
  );
};
