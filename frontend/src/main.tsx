import React from "react";

import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { Provider as JotaiProvider } from "jotai";
import { createRoot } from "react-dom/client";

import { ToastRegion } from "#/components/ui/ToastRegion/ToastRegion";
import { transport } from "#/lib/api/transport";
import { isTauri } from "#/lib/platform/platform";
import { router } from "#/lib/router";
import { queryClient } from "#/providers/query/query";
import { store } from "#/providers/store/store";
import { ThemeProvider } from "#/providers/theme/ThemeProvider";
import "@fontsource/ibm-plex-mono/400.css";
import "@fontsource/ibm-plex-mono/500.css";
import "@fontsource/ibm-plex-sans-jp/400.css";
import "@fontsource/ibm-plex-sans-jp/500.css";
import "@fontsource/ibm-plex-sans-jp/600.css";
import "@fontsource/ibm-plex-sans-jp/700.css";

import "./styles/globals.css";

if (isTauri) {
  const { interceptExternalLinks } = await import("#/lib/platform/tauri/externalLinks");
  interceptExternalLinks();
} else {
  const [{ listenInstallPrompt }, { registerServiceWorker }] = await Promise.all([
    import("#/features/layout/utils/installPrompt"),
    import("#/lib/registerServiceWorker"),
  ]);
  listenInstallPrompt();
  registerServiceWorker();
}

const rootEl = document.querySelector("#root");
if (rootEl) {
  createRoot(rootEl).render(
    <React.StrictMode>
      <JotaiProvider store={store}>
        <QueryClientProvider client={queryClient}>
          <TransportProvider transport={transport}>
            <ThemeProvider>
              <ToastRegion />
              <RouterProvider router={router} />
            </ThemeProvider>
          </TransportProvider>
        </QueryClientProvider>
      </JotaiProvider>
    </React.StrictMode>,
  );
}
