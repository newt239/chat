import React from "react";

import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { Provider as JotaiProvider } from "jotai";
import { createRoot } from "react-dom/client";

import { App } from "./App";
import { ToastRegion } from "./components/ui/ToastRegion";
import { transport } from "./lib/api/transport";
import { queryClient } from "./providers/query/query";
import { store } from "./providers/store";
import { ThemeProvider } from "./providers/theme/ThemeProvider";
import "@fontsource/ibm-plex-mono/400.css";
import "@fontsource/ibm-plex-mono/500.css";
import "@fontsource/ibm-plex-sans-jp/400.css";
import "@fontsource/ibm-plex-sans-jp/500.css";
import "@fontsource/ibm-plex-sans-jp/600.css";
import "@fontsource/ibm-plex-sans-jp/700.css";

import "./styles/globals.css";

const rootEl = document.querySelector("#root");
if (rootEl) {
  createRoot(rootEl).render(
    <React.StrictMode>
      <JotaiProvider store={store}>
        <QueryClientProvider client={queryClient}>
          <TransportProvider transport={transport}>
            <ThemeProvider>
              <ToastRegion />
              <App />
            </ThemeProvider>
          </TransportProvider>
        </QueryClientProvider>
      </JotaiProvider>
    </React.StrictMode>,
  );
}
