import { useApps } from "../hooks/useApps";
import { AppDialog } from "./AppDialog";

type AppDialogLoaderProps = {
  workspaceId: string;
  // null なら新規作成
  appId: string | null;
  initialChannelId: string | null;
  onClose: () => void;
};

// URL の ID から編集するアプリを引いてダイアログを開く。入力欄の初期値にするため、読み込めてから描く
export const AppDialogLoader = ({
  workspaceId,
  appId,
  initialChannelId,
  onClose,
}: AppDialogLoaderProps) => {
  const { data } = useApps(workspaceId);
  const app = data?.apps.find((candidate) => candidate.id === appId) ?? null;

  if (data === undefined || (appId !== null && (app === null || !app.canManage))) {
    return null;
  }
  return (
    <AppDialog
      key={appId}
      workspaceId={workspaceId}
      app={app}
      initialChannelId={initialChannelId}
      onClose={onClose}
    />
  );
};
