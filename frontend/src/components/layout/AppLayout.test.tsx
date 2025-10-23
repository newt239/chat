import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { AppLayout } from "./AppLayout";

import { queryClient } from "@/lib/query";

// Mock AuthGuard
vi.mock("@/features/auth/components/AuthGuard", () => ({
  AuthGuard: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="auth-guard">{children}</div>
  ),
}));

// Mock Header
vi.mock("./Header", () => ({
  Header: () => <div data-testid="header">Header</div>,
}));

const Wrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
};

describe("AppLayout", () => {
  it("renders with AuthGuard and Header", () => {
    render(
      <AppLayout>
        <div>Test Content</div>
      </AppLayout>,
      { wrapper: Wrapper }
    );

    expect(screen.getByTestId("auth-guard")).toBeInTheDocument();
    expect(screen.getByTestId("header")).toBeInTheDocument();
    expect(screen.getByText("Test Content")).toBeInTheDocument();
  });

  it("renders children in content area", () => {
    render(
      <AppLayout>
        <div data-testid="test-content">Test Content</div>
      </AppLayout>,
      { wrapper: Wrapper }
    );

    expect(screen.getByTestId("test-content")).toBeInTheDocument();
  });
});
