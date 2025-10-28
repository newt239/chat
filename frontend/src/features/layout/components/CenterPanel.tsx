import { Outlet, useParams } from "@tanstack/react-router";

import { ChannelHeader } from "@/features/channel/components/ChannelHeader";
import { MessageInput } from "@/features/message/components/MessageInput";

export const CenterPanel = () => {
  const params = useParams({ strict: false });
  const channelId = params.channelId as string;

  return (
    <div className="flex flex-col h-full min-h-0 bg-white">
      {/* CenterPanel ヘッダー */}
      <ChannelHeader channelId={channelId} />
      {/* メッセージ表示エリア */}
      <div className="flex-1 min-h-0">
        <Outlet />
      </div>

      {/* メッセージ入力エリア */}
      <MessageInput channelId={channelId} />
    </div>
  );
};
