import { ExpirationPlugin } from "workbox-expiration";
import {
  cleanupOutdatedCaches,
  createHandlerBoundToURL,
  precacheAndRoute,
} from "workbox-precaching";
import { NavigationRoute, registerRoute } from "workbox-routing";
import { CacheFirst, NetworkFirst } from "workbox-strategies";
import { z } from "zod";

import { notificationClickSchema } from "../src/lib/serviceWorkerMessage";

import type { NotificationClick } from "../src/lib/serviceWorkerMessage";

declare const self: ServiceWorkerGlobalScope;

// oxlint-disable-next-line no-underscore-dangle -- injectManifest がこの名前を置き換える
precacheAndRoute(self.__WB_MANIFEST);
cleanupOutdatedCaches();
// SPA なので、どのパスでも index.html を返す
registerRoute(new NavigationRoute(createHandlerBoundToURL("index.html")));

registerRoute(
  /\.woff2$/,
  new CacheFirst({ cacheName: "font-cache", plugins: [new ExpirationPlugin({ maxEntries: 500 })] }),
);
registerRoute(
  /^https:\/\/api\..*/i,
  new NetworkFirst({
    cacheName: "api-cache",
    plugins: [new ExpirationPlugin({ maxAgeSeconds: 60 * 60 * 24, maxEntries: 100 })],
  }),
);

// 更新のトーストで「再読み込み」が押されたら新しい版に切り替える
self.addEventListener("message", (event) => {
  if (z.object({ type: z.literal("SKIP_WAITING") }).safeParse(event.data).success) {
    void self.skipWaiting();
  }
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const parsed = notificationClickSchema.pick({ link: true }).safeParse(event.notification.data);
  if (!parsed.success) {
    return;
  }
  const { link } = parsed.data;
  event.waitUntil(
    (async () => {
      const [client] = await self.clients.matchAll({ includeUncontrolled: true, type: "window" });
      if (!client) {
        await self.clients.openWindow(link);
        return;
      }
      await client.focus();
      const message: NotificationClick = { link, type: "notification-click" };
      client.postMessage(message);
    })(),
  );
});
