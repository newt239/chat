import { useEffect, useState } from "react";

import { useAuthStore } from "@/lib/store/auth";

/**
 * 認証状態をチェックし、未認証の場合はログイン画面にリダイレクトするフック
 */
export function useAuthGuard() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const accessToken = useAuthStore((state) => state.accessToken);
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    // ストアの初期化を待つ
    const timer = setTimeout(() => {
      setIsInitialized(true);
    }, 100);

    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (!isInitialized) return;

    // 認証状態をより厳密にチェック
    const hasValidAuth = isAuthenticated && user && accessToken;

    if (!hasValidAuth) {
      const currentPath = window.location.pathname;
      if (currentPath !== "/login" && currentPath !== "/register") {
        // 認証情報をクリアしてからリダイレクト
        useAuthStore.getState().clearAuth();
        window.location.href = "/login";
      }
    }
  }, [isAuthenticated, user, accessToken, isInitialized]);

  return {
    isAuthenticated: isAuthenticated && !!user && !!accessToken,
    user,
    isLoading: !isInitialized,
  };
}
