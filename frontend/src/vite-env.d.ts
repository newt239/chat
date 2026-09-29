/// <reference types="vite-plus/client" />
/// <reference types="google.accounts" />
/// <reference types="vite-plugin-pwa/client" />

// グローバル宣言のマージには interface が必要
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_WS_URL?: string;
  // 未設定なら Google ログインのボタンを出さない
  readonly VITE_GOOGLE_OAUTH_CLIENT_ID?: string;
}

// PWA のインストールを促すイベント。Chromium 系だけが送る
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<{ outcome: "accepted" | "dismissed" }>;
}

interface WindowEventMap {
  beforeinstallprompt: BeforeInstallPromptEvent;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
