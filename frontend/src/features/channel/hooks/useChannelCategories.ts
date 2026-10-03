import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { ChannelCategoryService } from "#/gen/chat/v1/channel_category_service_pb";
import { toastError } from "#/lib/toastError";

import type { ListChannelCategoriesResponse } from "#/gen/chat/v1/channel_category_service_pb";

const channelCategoriesKey = (workspaceId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { workspaceId },
    schema: ChannelCategoryService.method.listChannelCategories,
  });

export const useChannelCategories = (workspaceId: string) =>
  useQuery(
    ChannelCategoryService.method.listChannelCategories,
    { workspaceId },
    { select: (res) => res.categories },
  );

/** 自分だけのサイドバーのカテゴリの作成・名前変更・削除・並び替えと、チャンネルの割り当て */
export const useChannelCategoryActions = (workspaceId: string) => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({ queryKey: channelCategoriesKey(workspaceId) });
  };
  const options = { onError: toastError, onSuccess };
  const create = useMutation(ChannelCategoryService.method.createChannelCategory, options);
  const update = useMutation(ChannelCategoryService.method.updateChannelCategory, options);
  const remove = useMutation(ChannelCategoryService.method.deleteChannelCategory, options);
  const setChannel = useMutation(ChannelCategoryService.method.setChannelCategory, options);
  const reorder = useMutation(ChannelCategoryService.method.reorderChannelCategories, {
    // 並びは押した直後に反映し、失敗したら取り直す
    onMutate: ({ categoryIds = [] }) => {
      queryClient.setQueriesData<ListChannelCategoriesResponse>(
        { queryKey: channelCategoriesKey(workspaceId) },
        (res) =>
          res && {
            ...res,
            categories: categoryIds.flatMap(
              (id) => res.categories.find((category) => category.id === id) ?? [],
            ),
          },
      );
    },
    onSettled: onSuccess,
  });

  return {
    create,
    move: (categoryIds: string[], from: number, to: number) => {
      const next = categoryIds.toSpliced(from, 1).toSpliced(to, 0, categoryIds[from] ?? "");
      reorder.mutate({ categoryIds: next, workspaceId });
    },
    remove,
    setChannel,
    update,
  };
};
