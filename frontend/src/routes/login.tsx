import { createFileRoute } from "@tanstack/react-router";

import { LoginForm } from "@/features/auth/components/LoginForm";

export const Route = createFileRoute("/login")({
  component: () => (
    <div className="flex h-full items-center justify-center bg-gray-50">
      <LoginForm />
    </div>
  ),
});
