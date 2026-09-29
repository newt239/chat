/// <reference types="vite-plus/client" />
/// <reference types="google.accounts" />

// グローバル宣言のマージには interface が必要
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_WS_URL?: string;
  // 未設定なら Google ログインのボタンを出さない
  readonly VITE_GOOGLE_OAUTH_CLIENT_ID?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
