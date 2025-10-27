import { RouterProvider } from "@tanstack/react-router";

import { router } from "@/lib/router";
import { AuthInitializer } from "@/providers/auth/AuthInitializer";
import { WebSocketEventHandler } from "@/providers/websocket/WebSocketEventHandler";
import { WebSocketProvider } from "@/providers/websocket/WebSocketProvider";

export const App = () => {
  return (
    <WebSocketProvider>
      <WebSocketEventHandler />
      <AuthInitializer />
      <RouterProvider router={router} />
    </WebSocketProvider>
  );
};
