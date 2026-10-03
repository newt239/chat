import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { OgpDataSchema } from "#/gen/chat/v1/message_pb";

import { OgpCard } from "./OgpCard";

describe("OgpCard", () => {
  test("サイト名・タイトル・説明を表示し、タイトルは新しいタブで開く", () => {
    render(
      <OgpCard
        url="https://example.com/post"
        ogp={create(OgpDataSchema, { description: "説明", siteName: "Example", title: "記事" })}
      />,
    );

    expect(screen.getByText("Example")).toBeInTheDocument();
    expect(screen.getByText("説明")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: "記事" });
    expect(link).toHaveAttribute("href", "https://example.com/post");
    expect(link).toHaveAttribute("target", "_blank");
    expect(screen.queryByRole("button")).toBeNull();
  });

  test("サイト名がなければホスト名を出し、外すボタンを押せる", async () => {
    const onRemove = vi.fn<() => void>();
    render(
      <OgpCard
        url="https://example.com/post"
        ogp={create(OgpDataSchema, { title: "記事" })}
        onRemove={onRemove}
      />,
    );

    expect(screen.getByText("example.com")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "プレビューを外す" }));
    expect(onRemove).toHaveBeenCalledOnce();
  });
});
