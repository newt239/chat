import { getCurrentWindow } from "@tauri-apps/api/window";

// 0 以下を渡すとバッジを消す。Windows では対応していないため失敗しても無視する
export const setBadgeCount = async (count: number) => {
  try {
    await getCurrentWindow().setBadgeCount(count > 0 ? count : undefined);
  } catch {
    // 何もしない
  }
};
