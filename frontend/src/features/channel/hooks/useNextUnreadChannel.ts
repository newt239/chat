import { useNavigate, useParams } from "@tanstack/react-router";

import { usePreferences } from "#/hooks/usePreferences";

import { nextUnreadId, sidebarOrder } from "../utils/sidebarOrder";
import { useChannels } from "./useChannel";
import { useChannelCategories } from "./useChannelCategories";
import { useDMs } from "./useDM";

/** サイドバーの並びで次の未読の会話と、そこへ移る操作。未読がなければ nextId は undefined */
export const useNextUnreadChannel = (workspaceId: string) => {
  const navigate = useNavigate();
  const { channelId } = useParams({ strict: false });
  const { channelSortOrder } = usePreferences();
  const { data: channels = [] } = useChannels(workspaceId);
  const { data: dms = [] } = useDMs(workspaceId);
  const { data: categories = [] } = useChannelCategories(workspaceId);
  const nextId = nextUnreadId(sidebarOrder(channels, dms, categories, channelSortOrder), channelId);

  const goNext = () => {
    if (nextId !== undefined) {
      void navigate({
        params: { channelId: nextId, workspaceId },
        to: "/app/$workspaceId/$channelId",
      });
    }
  };
  return { goNext, nextId };
};
