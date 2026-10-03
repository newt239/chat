import { UNSTABLE_ToastQueue as ToastQueue } from "react-aria-components";

type ToastTone = "default" | "success" | "danger";

type ToastAction = {
  label: string;
  onAction: () => void;
};

type ToastContent = {
  title: string;
  description?: string;
  tone: ToastTone;
  action?: ToastAction;
};

type ToastOptions = {
  description?: string;
  tone?: ToastTone;
  // 操作を付けたトーストは押すか閉じるまで出したままにする
  action?: ToastAction;
};

export const toastQueue = new ToastQueue<ToastContent>({ maxVisibleToasts: 3 });

// React のツリー外からも呼べる。表示は main.tsx の ToastRegion が担う
export const toast = (
  title: string,
  { action, description, tone = "default" }: ToastOptions = {},
) => toastQueue.add({ action, description, title, tone }, action ? {} : { timeout: 5000 });
