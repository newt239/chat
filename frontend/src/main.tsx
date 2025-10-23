import React from "react";
import { createRoot } from "react-dom/client";

import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { App } from "./App";
import { queryClient } from "./lib/query";
import "@mantine/core/styles.css";
import "./styles/globals.css";

const rootEl = document.getElementById("root");
if (rootEl) {
  createRoot(rootEl).render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <MantineProvider>
          <App />
          {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
        </MantineProvider>
      </QueryClientProvider>
    </React.StrictMode>
  );
}
