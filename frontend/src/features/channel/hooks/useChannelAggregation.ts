import { useAtom } from "jotai";

import { excludedDescendantsAtom } from "#/features/channel/atoms";

import { isDescendantPath } from "../utils/channelTree";
import { useChannels } from "./useChannel";

/** 親チャンネルで子孫のメッセージもまとめて表示するか。子孫がいれば既定でオン */
export const useChannelAggregation = (workspaceId: string, channelId: string) => {
  const { data: channels, isLoading, isError } = useChannels(workspaceId);
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
    isError,
    // チャンネル一覧を読み込むまでは集約するか決まらない
    isResolved: !isLoading,
    setIncludesDescendants: (value: boolean) => {
      setExcluded({ ...excluded, [channelId]: !value });
    },
  };
};
