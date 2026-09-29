import type { ReactNode } from "react";

import { useParams } from "@tanstack/react-router";

import { Link } from "#/components/ui/Link/Link";
import { cn } from "#/components/ui/styles/styles";
import { useChannels } from "#/features/channel/hooks/useChannel";

import { chipClassName } from "./chipClassName";

type ChannelLinkProps = {
  "data-channel": string;
  children?: ReactNode;
};

export const ChannelLink = ({ "data-channel": channelName }: ChannelLinkProps) => {
  const { workspaceId } = useParams({ strict: false });
  const { data: channels } = useChannels(workspaceId ?? null);
  const channel = channels?.find((item) => item.name === channelName);

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
