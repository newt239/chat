import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { CodeBlock } from "./CodeBlock";

const renderCodeBlock = (code: string, className?: string) =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <CodeBlock>
        <code className={className}>{code}</code>
      </CodeBlock>
    </QueryClientProvider>,
  );

describe("CodeBlock", () => {
  test("子の code 要素から言語を取り出し、行数とコピーボタンを表示する", () => {
    renderCodeBlock("const a = 1;\nconst b = 2;\n", "language-ts");

    expect(screen.getByText("ts")).toBeInTheDocument();
    expect(screen.getByText("2 行")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "コピー" })).toBeInTheDocument();
  });

  test("shiki でハイライトした結果に置き換える", async () => {
    const { container } = renderCodeBlock("const a = 1;", "language-ts");

    await waitFor(
      () => {
        expect(container.querySelector("pre.shiki")).toBeInTheDocument();
      },
      { timeout: 5000 },
    );
    expect(container.querySelector("pre.shiki")).toHaveTextContent("const a = 1;");
  });

  test("未対応の言語はプレーンテキストとして表示する", async () => {
    const { container } = renderCodeBlock("hello", "language-unknown");

    await waitFor(
      () => {
        expect(container.querySelector("pre.shiki")).toBeInTheDocument();
      },
      { timeout: 5000 },
    );
    expect(screen.getByText("unknown")).toBeInTheDocument();
  });

  test("12 行を超えるコードは折りたたみ、すべて表示で広げる", async () => {
    renderCodeBlock(Array.from({ length: 20 }, (_, i) => `line ${i}`).join("\n"));

    await userEvent.click(screen.getByRole("button", { name: "すべて表示（20 行）" }));
    expect(screen.queryByRole("button", { name: "すべて表示（20 行）" })).toBeNull();
  });

  test("短いコードは折りたたまない", () => {
    renderCodeBlock("a\nb");

    expect(screen.queryByRole("button", { name: /すべて表示/ })).toBeNull();
  });
});
