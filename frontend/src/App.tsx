import { RouterProvider } from "@tanstack/react-router";

import { router } from "@/lib/router";
import { AuthInitializer } from "@/providers/auth/AuthInitializer";

export const App = () => {
  return (
    <>
      <AuthInitializer />
      <RouterProvider router={router} />
    </>
  );
};
