import { RouterProvider } from "@tanstack/react-router";

import { router } from "@/lib/router";

export const App = () => {
  return <RouterProvider router={router} />;
};
