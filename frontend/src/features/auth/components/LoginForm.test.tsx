import type { ReactNode } from "react";

import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

vi.mock("@tanstack/react-router", () => ({
  Link: ({ to, children, ...rest }: { to: string; children: ReactNode }) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
}));

import { LoginForm } from "./LoginForm";

import { queryClient } from "@/providers/query/query";

// Mock the useLogin hook
vi.mock("../hooks/useAuth", () => ({
  useLogin: () => ({
    mutate: vi.fn(),
    isPending: false,
    isError: false,
    error: null,
  }),
}));

type WrapperProps = {
  children: ReactNode;
};

const Wrapper = ({ children }: WrapperProps) => {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
};

describe("LoginForm", () => {
  it("renders login form", () => {
    render(<LoginForm />, { wrapper: Wrapper });

    expect(screen.getByRole("heading", { name: "ログイン" })).toBeInTheDocument();
    expect(screen.getByLabelText(/メールアドレス/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/パスワード/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ログイン/i })).toBeInTheDocument();
  });

  it("allows input in form fields", async () => {
    const user = userEvent.setup();
    render(<LoginForm />, { wrapper: Wrapper });

    const emailInput = screen.getByLabelText(/メールアドレス/i);
    const passwordInput = screen.getByLabelText(/パスワード/i);

    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "password123");

    expect(emailInput).toHaveValue("test@example.com");
    expect(passwordInput).toHaveValue("password123");
  });

  it("shows link to registration page", () => {
    render(<LoginForm />, { wrapper: Wrapper });

    const registerLink = screen.getByRole("link", { name: /新規登録/i });
    expect(registerLink).toBeInTheDocument();
    expect(registerLink).toHaveAttribute("href", "/register");
  });

  it("shows validation errors when submitting empty fields", async () => {
    const user = userEvent.setup();
    render(<LoginForm />, { wrapper: Wrapper });

    const submitButton = screen.getByRole("button", { name: /ログイン/i });
    await user.click(submitButton);

    expect(screen.getByText("有効なメールアドレスを入力してください")).toBeInTheDocument();
    expect(screen.getByText("6文字以上のパスワードを入力してください")).toBeInTheDocument();
  });
});
