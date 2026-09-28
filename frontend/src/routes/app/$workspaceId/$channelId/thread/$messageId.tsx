import { createFileRoute } from "@tanstack/react-router";

import { store } from "#/providers/store";
import { closeRightSidePanelAtom, openThreadRouteAtom } from "#/providers/store/ui";

// 右パネル（モバイルでは全画面）にスレッドを開く。描画は AppShell の RightSidePanel が URL を見て行う
export const Route = createFileRoute("/app/$workspaceId/$channelId/thread/$messageId")({
  // 他のパネルを開いていても、スレッドを開いたらスレッドを表示する
  onEnter: ({ params }) => {
    store.set(closeRightSidePanelAtom);
    store.set(openThreadRouteAtom, params);
  },
  onLeave: () => {
    store.set(openThreadRouteAtom, null);
  },
  onStay: ({ params }) => {
    store.set(openThreadRouteAtom, params);
  },
});
