// API サーバーのオリジン。未指定ならフロントと同じオリジンで配信している前提
export const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || window.location.origin;
