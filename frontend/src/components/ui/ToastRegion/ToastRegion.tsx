import { IconAlertCircle, IconCircleCheck, IconX } from "@tabler/icons-react";
import {
  Button,
  Text,
  UNSTABLE_Toast as Toast,
  UNSTABLE_ToastContent as ToastContent,
  UNSTABLE_ToastRegion as AriaToastRegion,
} from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles/styles";

import { toastQueue } from "./toast";

const toneIcons = {
  danger: <IconAlertCircle aria-hidden className="size-4 shrink-0 text-danger" />,
  default: null,
  success: <IconCircleCheck aria-hidden className="size-4 shrink-0 text-success" />,
};

export const ToastRegion = () => {
  const { t } = useTranslation();
  return (
    <AriaToastRegion
      queue={toastQueue}
      aria-label={t("ui.toast.region")}
      className="fixed bottom-5 left-1/2 max-md:bottom-[calc(76px+max(10px,env(safe-area-inset-bottom)))] z-toast flex -translate-x-1/2 flex-col items-center gap-1.5 outline-none"
    >
      {({ toast }) => (
        <Toast
          toast={toast}
          className={`flex w-max max-w-[min(480px,calc(100vw-32px))] animate-pop-in items-center gap-2 rounded-md bg-text py-1.75 pr-1.5 pl-3.5 font-sans text-body-sm text-surface shadow-lg motion-reduce:animate-none ${focusRing}`}
        >
          {toneIcons[toast.content.tone]}
          <ToastContent className="flex min-w-0 flex-1 flex-col">
            <Text slot="title" className="font-semibold">
              {toast.content.title}
            </Text>
            {toast.content.description && (
              <Text slot="description" className="opacity-80">
                {toast.content.description}
              </Text>
            )}
          </ToastContent>
          {toast.content.action && (
            <Button
              onPress={() => {
                toast.content.action?.onAction();
                toastQueue.close(toast.key);
              }}
              className={`shrink-0 cursor-pointer rounded-sm px-2 py-1 whitespace-nowrap font-semibold text-accent-soft data-hovered:underline ${focusRing}`}
            >
              {toast.content.action.label}
            </Button>
          )}
          <Button
            slot="close"
            aria-label={t("ui.toast.dismiss")}
            className="grid size-6 shrink-0 cursor-pointer max-md:size-11 place-items-center rounded-sm opacity-70 outline-none data-focus-visible:opacity-100 data-hovered:opacity-100"
          >
            <IconX aria-hidden className="size-3.5" />
          </Button>
        </Toast>
      )}
    </AriaToastRegion>
  );
};
