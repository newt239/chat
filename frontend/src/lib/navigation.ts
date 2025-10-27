import { router } from "./router";

export const navigateToWorkspace = (workspaceId: string) => {
  router.navigate({
    to: "/app/$workspaceId",
    params: { workspaceId },
  });
};

export const navigateToChannel = (workspaceId: string, channelId: string) => {
  router.navigate({
    to: "/app/$workspaceId/$channelId",
    params: { workspaceId, channelId },
  });
};

export const navigateToLogin = () => {
  router.navigate({ to: "/login" });
};

export const navigateToAppWithWorkspace = () => {
  // ローカルストレージからワークスペース情報を取得
  const workspaceStorage = localStorage.getItem("workspace-storage");

  if (workspaceStorage) {
    try {
      const parsed = JSON.parse(workspaceStorage);
      const currentWorkspaceId = parsed.state?.currentWorkspaceId;

      if (currentWorkspaceId) {
        // ワークスペースが選択されている場合はそのページにリダイレクト
        navigateToWorkspace(currentWorkspaceId);
        return;
      }
    } catch (error) {
      console.warn("ワークスペース情報の解析に失敗しました:", error);
    }
  }

  // ワークスペース情報がない場合は通常のアプリページにリダイレクト
  router.navigate({ to: "/app" });
};
