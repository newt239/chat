import { createContext, useContext } from "react";

import { useParams } from "@tanstack/react-router";

// スレッドのパネルの中でだけ true
export const ThreadPanelContext = createContext(false);

// スレッドの親はチャンネルとパネルの両方に出るため、メッセージのダイアログはパネル側で描く
export const useOwnsMessageOverlay = (messageId: string) => {
  const isInThreadPanel = useContext(ThreadPanelContext);
  const threadId = useParams({ select: (params) => params.messageId, strict: false });
  return isInThreadPanel || threadId !== messageId;
};
