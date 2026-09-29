import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { RecordingPreview } from "./RecordingPreview";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("RecordingPreview", () => {
  test("再生ボタンでその場の audio を鳴らし、位置と長さを表示する", async () => {
    const play = vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue();
    const { container } = render(<RecordingPreview url="blob:voice" durationSeconds={12} />);

    expect(screen.getByText("0:00 / 0:12")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "再生" }));
    expect(play).toHaveBeenCalled();

    const audio = container.querySelector("audio");
    if (audio === null) {
      throw new Error("audio がありません");
    }
    fireEvent.play(audio);
    Object.defineProperty(audio, "currentTime", { configurable: true, value: 5 });
    fireEvent.timeUpdate(audio);

    expect(screen.getByRole("button", { name: "一時停止" })).toBeInTheDocument();
    expect(screen.getByText("0:05 / 0:12")).toBeInTheDocument();
  });
});
