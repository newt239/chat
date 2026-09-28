import React from "react";

import { TransportProvider } from "@connectrpc/connect-query";
import { Notifications } from "@mantine/notifications";
import { QueryClientProvider } from "@tanstack/react-query";
import { Provider as JotaiProvider } from "jotai";
import { createRoot } from "react-dom/client";

import { App } from "./App";
import { ToastRegion } from "./components/ui/ToastRegion";
import { transport } from "./lib/api/transport";
import { queryClient } from "./providers/query/query";
import { store } from "./providers/store";
import { ThemeProvider } from "./providers/theme/ThemeProvider";
import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";

import "./styles/globals.css";

const rootEl = document.querySelector("#root");
if (rootEl) {
  createRoot(rootEl).render(
    <React.StrictMode>
      <JotaiProvider store={store}>
        <QueryClientProvider client={queryClient}>
          <TransportProvider transport={transport}>
            <ThemeProvider>
              <Notifications />
              <ToastRegion />
              <App />
            </ThemeProvider>
          </TransportProvider>
        </QueryClientProvider>
      </JotaiProvider>
    </React.StrictMode>,
  );
}
