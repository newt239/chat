import { RouterProvider } from "@tanstack/react-router";

import { router } from "@/lib/router";
import { AuthInitializer } from "@/providers/auth/AuthInitializer";
import { WsProvider } from "@/providers/ws/WsProvider";

export const App = () => {
  return (
    <>
      <AuthInitializer />
      <WsProvider>
        <RouterProvider router={router} />
      </WsProvider>
    </>
  );
};
