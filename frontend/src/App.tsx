import { RouterProvider } from "@tanstack/react-router";

import { AuthInitializer } from "@/features/auth/components/AuthInitializer";
import { router } from "@/lib/router";

export const App = () => {
  return (
    <>
      <AuthInitializer />
      <RouterProvider router={router} />
    </>
  );
};
