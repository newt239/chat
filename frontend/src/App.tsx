import { useState, useEffect } from "react";

import { LoginForm } from "@/features/auth/components/LoginForm";
import { RegisterForm } from "@/features/auth/components/RegisterForm";
import { WorkspaceList } from "@/features/workspace/components/WorkspaceList";
import { useAuthStore } from "@/lib/store/auth";

type Page = "login" | "register" | "app";

export const App = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const [currentPage, setCurrentPage] = useState<Page>("login");

  useEffect(() => {
    const path = window.location.pathname;
    if (path === "/register") {
      setCurrentPage("register");
    } else if (path === "/app" || path.startsWith("/app/")) {
      setCurrentPage("app");
    } else {
      setCurrentPage("login");
    }
  }, []);

  useEffect(() => {
    if (isAuthenticated && currentPage !== "app") {
      setCurrentPage("app");
      window.history.pushState({}, "", "/app");
    } else if (!isAuthenticated && currentPage === "app") {
      setCurrentPage("login");
      window.history.pushState({}, "", "/login");
    }
  }, [isAuthenticated, currentPage]);

  if (currentPage === "register") {
    return (
      <div className="flex h-full items-center justify-center bg-gray-50">
        <RegisterForm />
      </div>
    );
  }

  if (currentPage === "login" || !isAuthenticated) {
    return (
      <div className="flex h-full items-center justify-center bg-gray-50">
        <LoginForm />
      </div>
    );
  }

  return (
    <div className="h-full p-6">
      <h1 className="text-2xl font-bold mb-4">ワークスペース</h1>
      <WorkspaceList />
    </div>
  );
}
