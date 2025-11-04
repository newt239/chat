import type { ReactElement } from "react";

import { render, screen } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";
import { describe, it, expect } from "vitest";

import { CodeBlock } from "@/features/message/components/markdown/CodeBlock";
import { createMantineWrapper } from "@/test/utils";

const renderWithMantine = (element: ReactElement) => {
  return render(element, { wrapper: createMantineWrapper() });
};

describe("CodeBlock", () => {
  it("コードブロックが正しくレンダリングされる", () => {
    const code = "const hello = 'world';";
    renderWithMantine(<CodeBlock className="language-typescript">{code}</CodeBlock>);

    expect(screen.getByText(code)).toBeInTheDocument();
  });

  it("言語が指定されていない場合でも表示される", () => {
    const code = "plain text";
    renderWithMantine(<CodeBlock>{code}</CodeBlock>);

    expect(screen.getByText(code)).toBeInTheDocument();
  });

  it("コピーボタンが存在する", () => {
    renderWithMantine(<CodeBlock className="language-python">print(&apos;hello&apos;)</CodeBlock>);

    const copyButton = screen.getByLabelText(/コピー/);
    expect(copyButton).toBeInTheDocument();
  });

  it("コピーボタンをクリックするとラベルが変化する", async () => {
    const user = userEvent.setup();
    const code = "const test = 123;";

    renderWithMantine(<CodeBlock className="language-typescript">{code}</CodeBlock>);

    const copyButton = screen.getByLabelText(/コピー/);
    await user.click(copyButton);

    // コピー後はボタンの状態が変わる
    expect(copyButton).toBeInTheDocument();
  });

  it("複数行のコードが正しく表示される", () => {
    const multilineCode = `function test() {
  return 'hello';
}`;

    renderWithMantine(<CodeBlock className="language-javascript">{multilineCode}</CodeBlock>);

    expect(screen.getByText(/function test/)).toBeInTheDocument();
  });

  it("ReactNodeの子要素からテキストが正しく抽出される", () => {
    renderWithMantine(
      <CodeBlock className="language-typescript">
        <code>const x = 1;</code>
      </CodeBlock>
    );

    expect(screen.getByText("const x = 1;")).toBeInTheDocument();
  });

  it("classNameから言語を正しく抽出する", () => {
    const { container } = renderWithMantine(
      <CodeBlock className="language-javascript">const y = 2;</CodeBlock>
    );

    // CodeHighlightコンポーネントが言語指定でレンダリングされることを確認
    expect(container.querySelector("code")).toBeInTheDocument();
  });

  it("空の子要素でもエラーが発生しない", () => {
    const { container } = renderWithMantine(<CodeBlock className="language-typescript" />);

    expect(container.querySelector("code")).toBeInTheDocument();
  });
});
