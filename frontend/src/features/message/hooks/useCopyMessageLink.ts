import { useCallback } from "react";

import { notifications } from "@mantine/notifications";
import { useRouter } from "@tanstack/react-router";

export const useCopyMessageLink = (workspaceId: string | null, channelId: string | null) => {
  const router = useRouter();

  return useCallback(
    (messageId: string) => {
      if (workspaceId === null || channelId === null) {
        return;
      }
      const { href } = router.buildLocation({
        params: { channelId, workspaceId },
        search: { message: messageId },
        to: "/app/$workspaceId/$channelId",
      });
      // 共有用のリンクなので origin を含む絶対 URL にする
      void navigator.clipboard.writeText(`${window.location.origin}${href}`);
      notifications.show({
        message: "メッセージリンクをクリップボードにコピーしました",
        title: "コピーしました",
      });
    },
    [router, workspaceId, channelId],
  );
};
