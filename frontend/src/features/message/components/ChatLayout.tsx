import { useAtomValue } from "jotai";

import { ChannelList } from "@/features/channel/components/ChannelList";
import { MessagePanel } from "@/features/message/components/MessagePanel";
import { MemberPanel } from "@/features/workspace/components/MemberPanel";
import { isMemberPanelOpenAtom } from "@/lib/store/ui";

interface ChatLayoutProps {
  workspaceId: string;
  channelId: string | null;
}

export const ChatLayout = ({ workspaceId, channelId }: ChatLayoutProps) => {
  const isMemberPanelOpen = useAtomValue(isMemberPanelOpenAtom);

  return (
    <div
      className={`grid h-full min-h-0 gap-6 ${
        isMemberPanelOpen ? "lg:grid-cols-[320px_1fr_280px]" : "lg:grid-cols-[320px_1fr]"
      }`}
    >
      <div className="space-y-6">
        <ChannelList workspaceId={workspaceId} />
      </div>
      <MessagePanel workspaceId={workspaceId} channelId={channelId} />
      <MemberPanel workspaceId={workspaceId} isOpen={isMemberPanelOpen} />
    </div>
  );
};
