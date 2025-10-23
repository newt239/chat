import type { ReactNode } from "react";

import { useAuthGuard } from "../hooks/useAuthGuard";

interface AuthGuardProps {
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * 認証が必要なコンポーネントを保護するガードコンポーネント
 * 未認証の場合は自動的にログイン画面にリダイレクトする
 */
export const AuthGuard = ({ children, fallback }: AuthGuardProps) => {
  const { isAuthenticated, isLoading } = useAuthGuard();

  // ローディング中の場合はフォールバックを表示
  if (isLoading) {
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
    return null;
  }

  // 認証済みの場合は子コンポーネントを表示
  return <>{children}</>;
};
