import { useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { copyWithToast } from "#/lib/clipboard";
import { toShareUrl } from "#/lib/platform/appOrigin";

export const useCopyMessageLink = (workspaceId: string | null, channelId: string | null) => {
  const router = useRouter();
  const { t } = useTranslation();

  return (messageId: string) => {
    if (workspaceId === null || channelId === null) {
      return;
    }
    const { href } = router.buildLocation({
      params: { channelId, workspaceId },
      search: { message: messageId },
      to: "/app/$workspaceId/$channelId",
    });
    void copyWithToast(toShareUrl(href), t("message.link.copied"));
  };
};
