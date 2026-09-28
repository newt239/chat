import { createContext, useContext } from "react";

import { useParams } from "@tanstack/react-router";

// スレッドのパネルの中でだけ true
export const ThreadPanelContext = createContext(false);

// メッセージに紐づくダイアログ（リアクション一覧・ライトボックス・操作シート）をこの MessageItem で描くか。
// スレッドの親メッセージはチャンネルとスレッドのパネルの両方に出るため、パネルの側で描く
export const useOwnsMessageOverlay = (messageId: string) => {
  const isInThreadPanel = useContext(ThreadPanelContext);
  const threadId = useParams({ select: (params) => params.messageId, strict: false });
  return isInThreadPanel || threadId !== messageId;
};
