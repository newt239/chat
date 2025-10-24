import type { ReactNode } from "react";

import { useAuthGuard } from "../hooks/useAuthGuard";

type AuthGuardProps = {
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * 認証が必要なコンポーネントを保護するガードコンポーネント
 * 未認証の場合は自動的にログイン画面にリダイレクトする
 */
export const AuthGuard = ({ children, fallback }: AuthGuardProps) => {
  const { isAuthenticated, isLoading } = useAuthGuard();

  // デバッグログを追加
  console.log("AuthGuard - レンダリング状態:", { isAuthenticated, isLoading });

  // ローディング中の場合はフォールバックを表示
  if (isLoading) {
    console.log("AuthGuard - ローディング中");
    return (
      fallback || (
        <div className="flex h-full items-center justify-center bg-gray-50">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-2 text-gray-600">認証状態を確認中...</p>
          </div>
        </div>
      )
    );
  }

  // 未認証の場合は何も表示しない（useAuthGuardでリダイレクトされる）
  if (!isAuthenticated) {
    console.log("AuthGuard - 未認証、何も表示しない");
    return null;
  }

  // 認証済みの場合は子コンポーネントを表示
  console.log("AuthGuard - 認証済み、子コンポーネントを表示");
  return <>{children}</>;
};
