import { toast } from "#/components/ui/ToastRegion/toast";
import { i18n } from "#/lib/i18n";

import type { ConnectError } from "@connectrpc/connect";

// サーバーのメッセージがなければ共通の文言を出す
export const toastError = (error: ConnectError) => {
  toast(error.rawMessage || i18n.t("common.actionFailed"), { tone: "danger" });
};
