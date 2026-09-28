import { useAtom } from "jotai";

import { excludedDescendantsAtom } from "#/providers/store/ui";

import { isDescendantPath } from "../utils/channelTree";
import { useChannels } from "./useChannel";

/** 親チャンネルで子孫のメッセージもまとめて表示するか。子孫がいれば既定でオン */
export const useChannelAggregation = (workspaceId: string | null, channelId: string | null) => {
  const { data: channels, isLoading } = useChannels(workspaceId);
  const [excluded, setExcluded] = useAtom(excludedDescendantsAtom);
  const channel = channels?.find((candidate) => candidate.id === channelId);
  const descendants = channel
    ? (channels ?? []).filter((candidate) => isDescendantPath(channel.name, candidate.name))
    : [];
  const includesDescendants =
    channel !== undefined && descendants.length > 0 && !(excluded[channel.id] ?? false);

  return {
    channel,
    descendants,
    includesDescendants,
    // チャンネル一覧を読み込むまでは集約するか決まらない
    isResolved: !isLoading,
    setIncludesDescendants: (value: boolean) => {
      if (channelId !== null) {
        setExcluded({ ...excluded, [channelId]: !value });
      }
    },
  };
};
