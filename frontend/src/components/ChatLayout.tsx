import { ChannelList } from "@/features/channel/components/ChannelList";
import { MessagePanel } from "@/features/message/components/MessagePanel";

interface ChatLayoutProps {
  workspaceId: string;
  channelId: string | null;
}

export const ChatLayout = ({ workspaceId, channelId }: ChatLayoutProps) => {
  return (
    <div className="grid gap-6 lg:grid-cols-[320px_1fr]">
      <div className="space-y-6">
        <ChannelList workspaceId={workspaceId} />
      </div>
      <MessagePanel workspaceId={workspaceId} channelId={channelId} />
    </div>
  );
};
