import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClientProvider } from "@tanstack/react-query";
import { MantineProvider } from "@mantine/core";
import { RegisterForm } from "./RegisterForm";
import { queryClient } from "@/lib/query";

// Mock the useRegister hook
vi.mock("../hooks/useAuth", () => ({
  useRegister: () => ({
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

describe("RegisterForm", () => {
  it("renders registration form", () => {
    render(<RegisterForm />, { wrapper: Wrapper });

    expect(screen.getByText("新規登録")).toBeInTheDocument();
    expect(screen.getByLabelText(/表示名/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/メールアドレス/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/パスワード/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /登録/i })).toBeInTheDocument();
  });

  it("allows input in form fields", async () => {
    const user = userEvent.setup();
    render(<RegisterForm />, { wrapper: Wrapper });

    const displayNameInput = screen.getByLabelText(/表示名/i);
    const emailInput = screen.getByLabelText(/メールアドレス/i);
    const passwordInput = screen.getByLabelText(/パスワード/i);

    await user.type(displayNameInput, "Test User");
    await user.type(emailInput, "test@example.com");
    await user.type(passwordInput, "password123");

    expect(displayNameInput).toHaveValue("Test User");
    expect(emailInput).toHaveValue("test@example.com");
    expect(passwordInput).toHaveValue("password123");
  });

  it("shows link to login page", () => {
    render(<RegisterForm />, { wrapper: Wrapper });

    const loginLink = screen.getByRole("link", { name: /ログイン/i });
    expect(loginLink).toBeInTheDocument();
    expect(loginLink).toHaveAttribute("href", "/login");
  });
});
