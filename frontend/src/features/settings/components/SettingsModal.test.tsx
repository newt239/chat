import { render, screen, fireEvent } from "@testing-library/react";
import { Provider } from "jotai";
import { vi } from "vitest";

import { SettingsModal } from "./SettingsModal";

import { MantineTestWrapper } from "@/test/MantineTestWrapper";

// router.navigateのモック
const mockNavigate = vi.fn();
vi.mock("@/lib/router", () => ({
  router: {
    navigate: mockNavigate,
  },
}));

// useSetAtomのモック
const mockClearAuth = vi.fn();
vi.mock("jotai", async () => {
  const actual = await vi.importActual("jotai");
  return {
    ...actual,
    useSetAtom: () => mockClearAuth,
  };
});

const TestWrapper = ({ children }: { children: React.ReactNode }) => (
  <Provider>
    <MantineTestWrapper>{children}</MantineTestWrapper>
  </Provider>
);

describe("SettingsModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("ログアウトボタンをクリックすると認証情報がクリアされ、ログインページにリダイレクトされる", () => {
    const mockOnClose = vi.fn();

    render(
      <TestWrapper>
        <SettingsModal opened={true} onClose={mockOnClose} />
      </TestWrapper>
    );

    const logoutButton = screen.getByRole("button", { name: "ログアウト" });
    fireEvent.click(logoutButton);

    expect(mockClearAuth).toHaveBeenCalled();
    expect(mockOnClose).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith({ to: "/login" });
  });

  it("キャンセルボタンをクリックするとモーダルが閉じられる", () => {
    const mockOnClose = vi.fn();

    render(
      <TestWrapper>
        <SettingsModal opened={true} onClose={mockOnClose} />
      </TestWrapper>
    );

    const cancelButton = screen.getByRole("button", { name: "キャンセル" });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });
});
