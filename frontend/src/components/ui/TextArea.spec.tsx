import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { TextArea } from "./TextArea";

describe("TextArea", () => {
  test("複数行の入力を通知する", async () => {
    const onChange = vi.fn<(value: string) => void>();
    render(<TextArea label="説明" rows={4} onChange={onChange} />);

    const textarea = screen.getByRole("textbox", { name: "説明" });
    expect(textarea.tagName).toBe("TEXTAREA");
    expect(textarea).toHaveAttribute("rows", "4");

    await userEvent.type(textarea, "a{Enter}b");

    expect(onChange).toHaveBeenLastCalledWith("a\nb");
  });

  test("errorMessage を表示する", () => {
    render(<TextArea label="説明" errorMessage="長すぎます" />);

    expect(screen.getByRole("textbox", { name: "説明" })).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByText("長すぎます")).toBeInTheDocument();
  });
});
