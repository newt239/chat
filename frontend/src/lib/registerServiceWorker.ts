import { registerSW } from "virtual:pwa-register";

import { toast } from "#/components/ui/toast";
import { i18n } from "#/lib/i18n";
import { navigateTo } from "#/lib/navigation";
import { notificationClickSchema } from "#/lib/serviceWorkerMessage";

// 新しい版を見つけたらトーストで知らせ、押されたら切り替えて再読み込みする
export const registerServiceWorker = () => {
  const updateServiceWorker = registerSW({
    onNeedRefresh: () => {
      toast(i18n.t("pwa.update.available"), {
        action: {
          label: i18n.t("pwa.update.reload"),
          onAction: () => {
            void updateServiceWorker(true);
          },
        },
      });
    },
  });

  // 通知を押したときに開いているタブを遷移させる
  if (!("serviceWorker" in navigator)) {
    return;
  }
  navigator.serviceWorker.addEventListener("message", (event) => {
    const link = notificationClickSchema.safeParse(event.data).data?.link;
    if (link !== undefined) {
      navigateTo({ href: link });
    }
  });
};
