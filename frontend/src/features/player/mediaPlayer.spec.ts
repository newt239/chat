import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { createMediaPlayer } from "./mediaPlayer";

const track = (attachmentId: string, kind: "audio" | "video") => ({
  attachmentId,
  authorName: "Alice",
  channelId: "channel",
  durationSeconds: 30,
  fileName: `${attachmentId}.mp4`,
  kind,
  messageId: "message",
  parentId: undefined,
  workspaceId: "workspace",
});

let intersect: (isIntersecting: boolean) => void = () => {};

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockReturnValue();
  vi.spyOn(HTMLMediaElement.prototype, "load").mockReturnValue();
  vi.stubGlobal(
    "IntersectionObserver",
    class {
      public constructor(callback: (entries: { isIntersecting: boolean }[]) => void) {
        intersect = (isIntersecting) => {
          callback([{ isIntersecting }]);
        };
      }
      public observe() {}
      public disconnect() {}
    },
  );
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("createMediaPlayer", () => {
  test("再生するのは常に 1 つだけで、別の添付を再生すると入れ替わる", async () => {
    const player = createMediaPlayer();
    await player.play(track("a", "audio"), () => Promise.resolve("https://example.com/a"));
    expect(player.getState()).toMatchObject({ isPlaying: true, track: { attachmentId: "a" } });

    await player.play(track("b", "audio"), () => Promise.resolve("https://example.com/b"));
    expect(player.getState().track?.attachmentId).toBe("b");

    player.stop();
    expect(player.getState()).toMatchObject({ isPlaying: false, position: 0, track: null });
  });

  test("URL を取得できなければ停止して例外を投げる", async () => {
    const player = createMediaPlayer();
    await expect(
      player.play(track("a", "audio"), () => Promise.reject(new Error("forbidden"))),
    ).rejects.toThrow("forbidden");
    expect(player.getState().track).toBeNull();
  });

  test("動画は見えているメッセージの枠、なければミニプレイヤーに映し、枠がなくなっても DOM に残す", async () => {
    const player = createMediaPlayer();
    const inline = document.createElement("div");
    const mini = document.createElement("div");
    document.body.append(inline, mini);
    await player.play(track("v", "video"), () => Promise.resolve("https://example.com/v"));
    const video = document.querySelector("video");

    const unregisterInline = player.registerSlot("inline", inline);
    const detach = player.attachInline(inline);
    intersect(true);
    expect(player.getState().isInlineVisible).toBe(true);
    expect(video?.parentElement).toBe(inline);

    const unregisterMini = player.registerSlot("mini", mini);
    expect(video?.parentElement).toBe(inline);
    intersect(false);
    expect(video?.parentElement).toBe(mini);

    detach();
    unregisterInline();
    unregisterMini();
    expect(video?.isConnected).toBe(true);
    expect(video?.parentElement).not.toBe(mini);
    player.stop();
  });

  test("元のメッセージが描画されていればスクロールし、なければ false を返す", () => {
    const player = createMediaPlayer();
    const inline = document.createElement("div");
    const scrollIntoView = vi.fn<() => void>();
    inline.scrollIntoView = scrollIntoView;
    expect(player.scrollToSource()).toBe(false);

    const detach = player.attachInline(inline);
    expect(player.scrollToSource()).toBe(true);
    expect(scrollIntoView).toHaveBeenCalledOnce();
    detach();
    expect(player.scrollToSource()).toBe(false);
  });

  test("再生速度は 1x → 1.5x → 2x → 1x の順に切り替わる", () => {
    const player = createMediaPlayer();
    const rates = [1, 2, 3].map(() => {
      player.cycleRate();
      return player.getState().rate;
    });
    player.cycleRate();
    expect([...rates, player.getState().rate]).toEqual([1.5, 2, 1, 1.5]);
  });
});
