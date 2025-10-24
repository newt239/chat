import { MessagePanel } from "@/features/message/components/MessagePanel";

interface ChatLayoutProps {
  workspaceId: string;
  channelId: string | null;
}

export const ChatLayout = ({ workspaceId, channelId }: ChatLayoutProps) => {
  return <MessagePanel workspaceId={workspaceId} channelId={channelId} />;
};
