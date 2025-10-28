import React from "react";
import { createRoot } from "react-dom/client";

import { MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import { QueryClientProvider } from "@tanstack/react-query";
import { Provider as JotaiProvider } from "jotai";

import { App } from "./App";
import { queryClient } from "./providers/query/query";
import { store } from "./providers/store";
import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";
import "./styles/globals.css";

const rootEl = document.getElementById("root");
if (rootEl) {
  createRoot(rootEl).render(
    <React.StrictMode>
      <JotaiProvider store={store}>
        <QueryClientProvider client={queryClient}>
          <MantineProvider>
            <Notifications />
            <App />
          </MantineProvider>
        </QueryClientProvider>
      </JotaiProvider>
    </React.StrictMode>
  );
}
