import type { ReactNode } from "react";

import { Badge } from "@mantine/core";
import { useNavigate } from "react-router";

import { paths } from "#/lib/paths";
import { useOptionalRouteParams } from "#/lib/routeParams";

type ChannelLinkProps = {
  "data-channel": string;
  children?: ReactNode;
};

export const ChannelLink = ({ "data-channel": channelName }: ChannelLinkProps) => {
  const navigate = useNavigate();
  const { workspaceId } = useOptionalRouteParams();

  const handleClick = () => {
    if (workspaceId === undefined) {
      return;
    }
    // チャンネル名からチャンネル ID を解決する必要がある
    // ここでは簡略化のため、チャンネル名をそのまま使用
    void navigate(paths.channel(workspaceId, channelName));
  };

  return (
    <Badge
      variant="light"
      color="green"
      size="sm"
      className="cursor-pointer hover:bg-green-100"
      component="span"
      onClick={handleClick}
    >
      #{channelName}
    </Badge>
  );
};
