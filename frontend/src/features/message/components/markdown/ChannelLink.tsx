import type { ReactNode } from "react";

import { Badge } from "@mantine/core";
import { useNavigate } from "react-router";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { paths } from "#/lib/paths";
import { useOptionalRouteParams } from "#/lib/routeParams";

type ChannelLinkProps = {
  "data-channel": string;
  children?: ReactNode;
};

export const ChannelLink = ({ "data-channel": channelName }: ChannelLinkProps) => {
  const navigate = useNavigate();
  const { workspaceId } = useOptionalRouteParams();
  const { data: channels } = useChannels(workspaceId ?? null);

  const channel = channels?.find((item) => item.name === channelName);

  const handleClick = () => {
    if (workspaceId === undefined || channel === undefined) {
      return;
    }
    void navigate(paths.channel(workspaceId, channel.id));
  };

  return (
    <Badge
      variant="light"
      color="green"
      size="sm"
      className={channel === undefined ? "" : "cursor-pointer hover:bg-green-100"}
      component="span"
      onClick={handleClick}
    >
      #{channelName}
    </Badge>
  );
};
