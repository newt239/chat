import { getCurrent, onOpenUrl } from "@tauri-apps/plugin-deep-link";

/** アプリを開いたディープリンクを受け取る。起動のきっかけになったリンクにも呼ぶ。戻り値で購読を止める */
export const listenDeepLinks = (handler: (url: string) => void) => {
  let active = true;
  const handle = (urls: string[] | null) => {
    if (!active) {
      return;
    }
    for (const url of urls ?? []) {
      handler(url);
    }
  };
  const unlisten = onOpenUrl(handle);
  void getCurrent().then(handle);
  return () => {
    active = false;
    void unlisten.then((stop) => {
      stop();
    });
  };
};
