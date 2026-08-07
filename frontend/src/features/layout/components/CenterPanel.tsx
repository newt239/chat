import { Outlet } from "react-router";

import { ChannelHeader } from "#/features/channel/components/ChannelHeader";
import { MessageInput } from "#/features/message/components/MessageInput";
import { useOptionalRouteParams } from "#/lib/routeParams";

export const CenterPanel = () => {
  const { channelId } = useOptionalRouteParams();

  return (
    <div className="flex flex-col h-full min-h-0 bg-white">
      {/* CenterPanel ヘッダー（チャンネル選択時のみ） */}
      {channelId !== undefined && <ChannelHeader channelId={channelId} />}
      {/* メッセージ表示エリア */}
      <div className="flex-1 min-h-0">
        <Outlet />
      </div>

      {/* メッセージ入力エリア（チャンネル選択時のみ） */}
      {channelId !== undefined && <MessageInput channelId={channelId} />}
    </div>
  );
};
