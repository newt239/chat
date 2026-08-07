import { RouterProvider } from "react-router";

import { router } from "#/lib/router";
import { WsProvider } from "#/providers/ws/WsProvider";

export const App = () => (
  <WsProvider>
    <RouterProvider router={router} />
  </WsProvider>
);
