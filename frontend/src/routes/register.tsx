import { createFileRoute } from "@tanstack/react-router";

import { RegisterForm } from "@/features/auth/components/RegisterForm";

export const Route = createFileRoute("/register")({
  component: () => (
    <div className="flex h-full items-center justify-center bg-gray-50">
      <RegisterForm />
    </div>
  ),
});
