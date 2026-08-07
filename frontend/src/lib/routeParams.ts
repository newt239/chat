import { useParams } from "react-router";

/** `/app/:workspaceId` 配下でのみ使用する。 ルート定義上 workspaceId は必ず存在するため、欠落は実装バグとして throw する。 */
export const useWorkspaceId = () => {
  const { workspaceId } = useParams<"workspaceId">();
  if (workspaceId === undefined) {
    throw new Error("workspaceId がルートパラメータに存在しません");
  }
  return workspaceId;
};

/** `/app/:workspaceId/:channelId` でのみ使用する。 */
export const useChannelId = () => {
  const { channelId } = useParams<"channelId">();
  if (channelId === undefined) {
    throw new Error("channelId がルートパラメータに存在しません");
  }
  return channelId;
};

/** ルート階層をまたぐ共通 UI 向け。パラメータが無い階層でも呼べる。 */
export const useOptionalRouteParams = () => useParams<"workspaceId" | "channelId">();
