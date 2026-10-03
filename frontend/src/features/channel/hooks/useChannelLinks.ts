import { createConnectQueryKey, skipToken, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { ChannelLinkService } from "#/gen/chat/v1/channel_link_service_pb";
import { toastError } from "#/lib/toastError";

import type { ListChannelLinksResponse } from "#/gen/chat/v1/channel_link_service_pb";

const channelLinksKey = (channelId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { channelId },
    schema: ChannelLinkService.method.listChannelLinks,
  });

export const useChannelLinks = (channelId: string | null) =>
  useQuery(
    ChannelLinkService.method.listChannelLinks,
    channelId === null ? skipToken : { channelId },
  );

/** 関連リンクの追加・編集・削除・並び替え */
export const useChannelLinkActions = (channelId: string) => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({ queryKey: channelLinksKey(channelId) });
  };
  const create = useMutation(ChannelLinkService.method.createChannelLink, { onSuccess });
  const update = useMutation(ChannelLinkService.method.updateChannelLink, { onSuccess });
  const remove = useMutation(ChannelLinkService.method.deleteChannelLink, {
    onError: toastError,
    onSuccess,
  });
  const reorder = useMutation(ChannelLinkService.method.reorderChannelLinks, {
    // 並びは押した直後に反映し、失敗したら取り直す
    onMutate: ({ linkIds = [] }) => {
      queryClient.setQueriesData<ListChannelLinksResponse>(
        { queryKey: channelLinksKey(channelId) },
        (res) =>
          res && {
            ...res,
            links: linkIds.flatMap((id) => res.links.find((link) => link.id === id) ?? []),
          },
      );
    },
    onSettled: onSuccess,
  });

  return {
    create,
    move: (linkIds: string[], from: number, to: number) => {
      const next = linkIds.toSpliced(from, 1).toSpliced(to, 0, linkIds[from] ?? "");
      reorder.mutate({ channelId, linkIds: next });
    },
    remove,
    update,
  };
};
