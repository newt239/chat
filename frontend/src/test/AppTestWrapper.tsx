import type { ReactNode } from "react";

import { MantineProvider } from "@mantine/core";
import { QueryClientProvider, type QueryClient } from "@tanstack/react-query";
import { Provider as JotaiProvider } from "jotai";

import { store } from "@/lib/store";

type AppTestWrapperProps = {
  children: ReactNode;
  queryClient: QueryClient;
};

export const AppTestWrapper = ({ children, queryClient }: AppTestWrapperProps) => {
  return (
    <JotaiProvider store={store}>
      <QueryClientProvider client={queryClient}>
        <MantineProvider>{children}</MantineProvider>
      </QueryClientProvider>
    </JotaiProvider>
  );
};
