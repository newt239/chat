import { useParams } from "react-router";

export const useWorkspaceId = () => {
  const { workspaceId } = useParams<"workspaceId">();
  if (workspaceId === undefined) {
    throw new Error("workspaceId がルートパラメータに存在しません");
  }
  return workspaceId;
};

export const useChannelId = () => {
  const { channelId } = useParams<"channelId">();
  if (channelId === undefined) {
    throw new Error("channelId がルートパラメータに存在しません");
  }
  return channelId;
};

export const useOptionalRouteParams = () => useParams<"workspaceId" | "channelId">();
