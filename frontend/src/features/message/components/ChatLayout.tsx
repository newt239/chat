import { useAtomValue } from "jotai";

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
    <div className="grid h-full min-h-0 gap-6 lg:grid-cols-[1fr_280px]">
      <MessagePanel workspaceId={workspaceId} channelId={channelId} />
      <MemberPanel workspaceId={workspaceId} isOpen={isMemberPanelOpen} />
    </div>
  );
};
