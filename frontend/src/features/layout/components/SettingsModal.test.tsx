import { render, screen, fireEvent } from "@testing-library/react";
import { Provider } from "jotai";
import { vi } from "vitest";

import { SettingsModal } from "./SettingsModal";

import * as navigation from "@/lib/navigation";
import { MantineTestWrapper } from "@/test/MantineTestWrapper";

// モック設定
vi.mock("@/lib/navigation", () => ({
  navigateToLogin: vi.fn(),
}));

const mockNavigateToLogin = vi.mocked(navigation.navigateToLogin);

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
    expect(mockNavigateToLogin).toHaveBeenCalled();
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
