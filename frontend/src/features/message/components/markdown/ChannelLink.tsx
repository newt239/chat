import type { ReactNode } from 'react';

import { Badge } from '@mantine/core';
import { useNavigate, useParams } from '@tanstack/react-router';

type ChannelLinkProps = {
  'data-channel': string;
  children?: ReactNode;
}

export const ChannelLink = ({ 'data-channel': channelName }: ChannelLinkProps) => {
  const navigate = useNavigate();
  const { workspaceId } = useParams({ strict: false });

  const handleClick = () => {
    if (!workspaceId) return;
    // チャンネル名からチャンネル ID を解決する必要がある
    // ここでは簡略化のため、チャンネル名をそのまま使用
    navigate({
      to: '/app/$workspaceId/$channelId',
      params: { workspaceId, channelId: channelName },
    });
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
}
