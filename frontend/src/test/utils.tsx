import type { ReactNode } from "react";

import { QueryClient } from "@tanstack/react-query";

import { AppTestWrapper } from "./AppTestWrapper";
import { MantineTestWrapper } from "./MantineTestWrapper";

type WrapperProps = {
  children: ReactNode;
};

const createDefaultQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

export const createMantineWrapper = () => MantineTestWrapper;

export const createAppWrapper = (queryClient?: QueryClient) => {
  const client = queryClient ?? createDefaultQueryClient();

  const Wrapper = ({ children }: WrapperProps) => {
    return <AppTestWrapper queryClient={client}>{children}</AppTestWrapper>;
  };

  Wrapper.displayName = "AppTestWrapperWithClient";

  return Wrapper;
};

export const createTestQueryClient = () => createDefaultQueryClient();
