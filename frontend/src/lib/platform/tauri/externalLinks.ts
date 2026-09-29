import { openUrl } from "@tauri-apps/plugin-opener";

import { navigateTo } from "#/lib/navigation";

const externalProtocols = new Set(["http:", "https:", "mailto:"]);

// WebView は新しいウィンドウを開けないため、外部のリンクはブラウザへ、アプリ内のリンクは同じ画面で開く
const handleClick = (event: MouseEvent) => {
  if (!(event.target instanceof Element)) {
    return;
  }
  const anchor = event.target.closest("a[href]");
  if (!(anchor instanceof HTMLAnchorElement)) {
    return;
  }
  const url = new URL(anchor.href);
  if (url.origin === globalThis.location.origin) {
    if (anchor.target === "_blank") {
      event.preventDefault();
      navigateTo({ href: `${url.pathname}${url.search}${url.hash}` });
    }
    return;
  }
  if (externalProtocols.has(url.protocol)) {
    event.preventDefault();
    void openUrl(url.href);
  }
};

export const interceptExternalLinks = () => {
  document.addEventListener("click", handleClick, true);
  document.addEventListener("auxclick", handleClick, true);
};
