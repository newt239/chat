import { toast } from "#/components/ui/ToastRegion/toast";
import { i18n } from "#/lib/i18n";

// サーバーのメッセージがなければ共通の文言を出す
export const toastError = (error: Error) => {
  toast(error.message || i18n.t("common.actionFailed"), { tone: "danger" });
};
