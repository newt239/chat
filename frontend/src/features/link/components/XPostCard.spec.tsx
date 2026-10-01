import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { OgpDataSchema, XPostSchema } from "#/gen/chat/v1/message_pb";

import { XPostCard } from "./XPostCard";

const url = "https://x.com/jack/status/20";
const post = create(XPostSchema, { authorHandle: "jack", authorName: "jack" });

describe("XPostCard", () => {
  test("投稿者・本文を出し、押すと X を新しいタブで開く", () => {
    render(
      <XPostCard
        url={url}
        ogp={create(OgpDataSchema, {
          cardType: "summary",
          description: "just setting up my twttr",
          imageUrl: "https://pbs.twimg.com/avatar.jpg",
        })}
        post={post}
      />,
    );

    const link = screen.getByRole("link", { name: "X で開く" });
    expect(link).toHaveAttribute("href", url);
    expect(link).toHaveAttribute("target", "_blank");
    expect(screen.getByText("@jack")).toBeInTheDocument();
    expect(screen.getByText("just setting up my twttr")).toBeInTheDocument();
    expect(link.querySelector("img")).toHaveClass("rounded-full");
  });

  test("画像付きの投稿は画像を大きく出す", () => {
    render(
      <XPostCard
        url={url}
        ogp={create(OgpDataSchema, {
          cardType: "summary_large_image",
          imageUrl: "https://pbs.twimg.com/media.jpg",
        })}
        post={post}
      />,
    );

    const image = screen.getByRole("link", { name: "X で開く" }).querySelector("img");
    expect(image).toHaveAttribute("src", "https://pbs.twimg.com/media.jpg");
    expect(image).not.toHaveClass("rounded-full");
  });
});
