import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { VoiceRecorder } from "./VoiceRecorder";

const stopTrack = vi.fn<() => void>();

// iOS Safari と同じく mp4 だけに対応する MediaRecorder
class FakeRecorder extends EventTarget {
  public static isTypeSupported = (type: string) => type === "audio/mp4";
  public state = "inactive";
  public readonly mimeType = "audio/mp4";

  public start() {
    this.state = "recording";
  }

  public stop() {
    this.state = "inactive";
    this.dispatchEvent(
      Object.assign(new Event("dataavailable"), {
        data: new Blob(["voice"], { type: "audio/mp4" }),
      }),
    );
    this.dispatchEvent(new Event("stop"));
  }
}

const stubMicrophone = (
  getUserMedia: () => Promise<{ getTracks: () => { stop: () => void }[] }>,
) => {
  Object.defineProperty(navigator, "mediaDevices", { configurable: true, value: { getUserMedia } });
};

beforeEach(() => {
  vi.stubGlobal("MediaRecorder", FakeRecorder);
  Object.defineProperty(URL, "createObjectURL", { configurable: true, value: () => "blob:voice" });
  Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: () => {} });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("VoiceRecorder", () => {
  test("録音を止めて試聴し、m4a のファイルとして添付する", async () => {
    stubMicrophone(() => Promise.resolve({ getTracks: () => [{ stop: stopTrack }] }));
    const onAttach = vi.fn<(file: File, durationSeconds: number) => void>();
    render(<VoiceRecorder onAttach={onAttach} onDiscard={vi.fn<() => void>()} />);

    expect(await screen.findByText("録音中")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "停止" }));

    expect(stopTrack).toHaveBeenCalled();
    expect(document.querySelector("audio")).toHaveAttribute("src", "blob:voice");
    await userEvent.click(screen.getByRole("button", { name: "添付する" }));

    const [file] = onAttach.mock.calls[0] ?? [];
    expect(file?.type).toBe("audio/mp4");
    expect(file?.name).toMatch(/^voice-\d{8}-\d{6}\.m4a$/);
  });

  test("マイクが許可されなければ理由を出し、破棄で閉じられる", async () => {
    stubMicrophone(() => Promise.reject(new Error("denied")));
    const onDiscard = vi.fn<() => void>();
    render(
      <VoiceRecorder
        onAttach={vi.fn<(file: File, durationSeconds: number) => void>()}
        onDiscard={onDiscard}
      />,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent("マイクの利用が許可されていません");
    await userEvent.click(screen.getByRole("button", { name: "録音を破棄" }));
    expect(onDiscard).toHaveBeenCalled();
  });
});
