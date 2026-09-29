import { z } from "zod";

// 通知を押したとき、Service Worker が開いているタブへ遷移先のパスを渡すメッセージ
export const notificationClickSchema = z.object({
  link: z.string().startsWith("/"),
  type: z.literal("notification-click"),
});

export type NotificationClick = z.infer<typeof notificationClickSchema>;
