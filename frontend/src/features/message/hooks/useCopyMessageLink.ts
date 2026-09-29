import { useCallback } from "react";

import { useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { toShareUrl } from "#/lib/platform/appOrigin";

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
      navigator.clipboard.writeText(toShareUrl(href)).then(
        () => toast(t("message.link.copied"), { tone: "success" }),
        () => toast(t("message.link.copyFailed"), { tone: "danger" }),
      );
    },
    [router, workspaceId, channelId, t],
  );
};
