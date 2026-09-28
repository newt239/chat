import { createFileRoute } from "@tanstack/react-router";

// 右パネル（モバイルでは全画面）にスレッドを開く。描画は AppShell の RightSidePanel が URL を見て行う
export const Route = createFileRoute("/app/$workspaceId/$channelId/thread/$messageId")({});
