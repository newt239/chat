import type { ReactElement } from "react";

import { render } from "@testing-library/react";
import { describe, it, expect } from "vitest";

import { CodeBlock } from "@/features/message/components/markdown/CodeBlock";
import { createMantineWrapper } from "@/test/utils";

const renderWithMantine = (element: ReactElement) => {
  return render(element, { wrapper: createMantineWrapper() });
};

describe("CodeBlock", () => {
  it("コードブロックをレンダリングできる", () => {
    const { container } = renderWithMantine(
      <CodeBlock className="language-typescript">const x = 1;</CodeBlock>
    );

    const code = container.querySelector("code");
    expect(code).toBeDefined();
    expect(code).toHaveClass("language-typescript");
  });

  it("言語指定なしのプレーンテキストを処理できる", () => {
    const { container } = renderWithMantine(<CodeBlock>plain text</CodeBlock>);

    const code = container.querySelector("code");
    expect(code).toBeDefined();
    expect(code).toHaveClass("language-text");
  });
});
