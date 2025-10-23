import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClientProvider } from "@tanstack/react-query";
import { MantineProvider } from "@mantine/core";
import { LoginForm } from "./LoginForm";
import { queryClient } from "@/lib/query";

// Mock the useLogin hook
vi.mock("../hooks/useAuth", () => ({
  useLogin: () => ({
    mutate: vi.fn(),
    isPending: false,
    isError: false,
    error: null,
  }),
}));

function Wrapper({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
}

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
});
