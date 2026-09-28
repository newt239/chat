import { UNSTABLE_ToastQueue as ToastQueue } from "react-aria-components";

type ToastTone = "default" | "success" | "danger";

export type ToastContent = {
  title: string;
  description?: string;
  tone: ToastTone;
};

type ToastOptions = {
  description?: string;
  tone?: ToastTone;
};

export const toastQueue = new ToastQueue<ToastContent>({ maxVisibleToasts: 3 });

// React のツリー外からも呼べる。表示は main.tsx の ToastRegion が担う
export const toast = (title: string, { description, tone = "default" }: ToastOptions = {}) =>
  toastQueue.add({ description, title, tone }, { timeout: 5000 });
