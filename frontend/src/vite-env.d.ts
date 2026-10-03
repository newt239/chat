/// <reference types="vite-plus/client" />
/// <reference types="google.accounts" />
/// <reference types="vite-plugin-pwa/client" />

// グローバル宣言のマージには interface が必要
interface ImportMetaEnv {
  // 共有用リンクに使う Web 版の URL。未設定なら表示中の origin
  readonly VITE_PUBLIC_APP_URL?: string;
  readonly VITE_API_BASE_URL: string;
  readonly VITE_WS_URL: string;
  // 未設定なら Google ログインのボタンを出さない
  readonly VITE_GOOGLE_OAUTH_CLIENT_ID?: string;
  // すべて揃っていなければプッシュ通知を使わない
  readonly VITE_FIREBASE_API_KEY?: string;
  readonly VITE_FIREBASE_PROJECT_ID?: string;
  readonly VITE_FIREBASE_MESSAGING_SENDER_ID?: string;
  readonly VITE_FIREBASE_APP_ID?: string;
  readonly VITE_FIREBASE_VAPID_KEY?: string;
  // Tauri の CLI が渡す。Web 版では未定義
  readonly TAURI_ENV_PLATFORM?: string;
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
