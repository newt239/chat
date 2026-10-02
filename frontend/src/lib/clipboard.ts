import { toast } from "#/components/ui/ToastRegion/toast";
import { i18n } from "#/lib/i18n";

/** クリップボードに書き込み、書き込めたときだけ successMessage を通知する */
export const copyWithToast = (text: string, successMessage: string) =>
  navigator.clipboard.writeText(text).then(
    () => {
      toast(successMessage, { tone: "success" });
    },
    () => {
      toast(i18n.t("common.copyFailed"), { tone: "danger" });
    },
  );
