import { useTranslation } from "react-i18next";

import { ChannelRow } from "#/features/channel/components/ChannelRow";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { DMRow } from "#/features/dm/components/DMRow";
import { useDMs } from "#/features/dm/hooks/useDM";

import { SidebarSection } from "./SidebarSection";

type StarredSectionProps = {
  workspaceId: string;
};

// スターを付けたチャンネルと DM。1 件もなければ見出しごと出さない
export const StarredSection = ({ workspaceId }: StarredSectionProps) => {
  const { t } = useTranslation();
  const { data: channels = [] } = useChannels(workspaceId);
  const { data: dms = [] } = useDMs(workspaceId);
  const starredChannels = channels.filter((channel) => channel.isStarred && channel.isMember);
  const starredDMs = dms.filter((dm) => dm.isStarred);

  if (starredChannels.length === 0 && starredDMs.length === 0) {
    return null;
  }

  return (
    <SidebarSection id="starred" title={t("shell.sidebar.starred")} onAdd={null}>
      {starredChannels.map((channel) => (
        <ChannelRow key={channel.id} workspaceId={workspaceId} channel={channel} />
      ))}
      {starredDMs.map((dm) => (
        <DMRow key={dm.id} workspaceId={workspaceId} dm={dm} />
      ))}
    </SidebarSection>
  );
};
