import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { OgpDataSchema, YouTubeVideoSchema } from "#/gen/chat/v1/message_pb";

import { YouTubeCard } from "./YouTubeCard";

const url = "https://www.youtube.com/watch?v=abc";

describe("YouTubeCard", () => {
  test("サムネイル・再生時間・チャンネル名を出し、押すと YouTube を新しいタブで開く", () => {
    render(
      <YouTubeCard
        url={url}
        ogp={create(OgpDataSchema, { imageUrl: "https://i.ytimg.com/thumb.jpg", title: "動画" })}
        video={create(YouTubeVideoSchema, {
          channelName: "Acme",
          durationSeconds: 1458,
          videoId: "abc",
        })}
      />,
    );

    const play = screen.getByRole("link", { name: "YouTube で再生" });
    expect(play).toHaveAttribute("href", url);
    expect(play).toHaveAttribute("target", "_blank");
    expect(play.querySelector("img")).toHaveAttribute("src", "https://i.ytimg.com/thumb.jpg");
    expect(screen.getByText("24:18")).toBeInTheDocument();
    expect(screen.getByText("YouTube · Acme")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "動画" })).toHaveAttribute("href", url);
  });

  test("サムネイルがなければ動画 ID から作り、再生時間がなければ出さない", () => {
    render(
      <YouTubeCard
        url={url}
        ogp={create(OgpDataSchema, {})}
        video={create(YouTubeVideoSchema, { videoId: "abc" })}
      />,
    );

    expect(
      screen.getByRole("link", { name: "YouTube で再生" }).querySelector("img"),
    ).toHaveAttribute("src", "https://i.ytimg.com/vi/abc/hqdefault.jpg");
    expect(screen.queryByText(/\d+:\d{2}/)).toBeNull();
    expect(screen.getByRole("link", { name: url })).toBeInTheDocument();
  });
});
