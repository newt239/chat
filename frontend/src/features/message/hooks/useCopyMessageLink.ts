import { useCallback } from "react";

import { useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";

export const useCopyMessageLink = (workspaceId: string | null, channelId: string | null) => {
  const router = useRouter();
  const { t } = useTranslation();

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
      navigator.clipboard.writeText(`${window.location.origin}${href}`).then(
        () => toast(t("message.link.copied"), { tone: "success" }),
        () => toast(t("message.link.copyFailed"), { tone: "danger" }),
      );
    },
    [router, workspaceId, channelId, t],
  );
};
