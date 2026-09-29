import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { DraftSchema, DraftService } from "#/gen/chat/v1/draft_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { useDraftAutosave } from "./useDraftAutosave";

import type { DeleteDraftRequest, SaveDraftRequest } from "#/gen/chat/v1/draft_service_pb";

const Harness = () => {
  const { discard, initialBody, save } = useDraftAutosave("ch1", "m1");
  return (
    <>
      <output>{initialBody ?? "loading"}</output>
      <button
        type="button"
        onClick={() => {
          save("一文字目");
          save("二文字目");
        }}
      >
        save
      </button>
      <button type="button" onClick={discard}>
        discard
      </button>
    </>
  );
};

describe("useDraftAutosave", () => {
  test("保存済みの本文を返し、連続した入力は最後の本文だけをまとめて保存する", async () => {
    const saveDraft = vi.fn<(req: SaveDraftRequest) => void>();
    const deleteDraft = vi.fn<(req: DeleteDraftRequest) => void>();
    await renderWithProviders(<Harness />, "/app/ws1", (routes) => {
      routes.rpc(DraftService.method.getDraft, ({ parentId }) => ({
        draft: create(DraftSchema, { body: `前回の続き ${parentId ?? ""}`, channelId: "ch1" }),
      }));
      routes.rpc(DraftService.method.saveDraft, (req) => {
        saveDraft(req);
        return {};
      });
      routes.rpc(DraftService.method.deleteDraft, (req) => {
        deleteDraft(req);
        return {};
      });
    });

    expect(await screen.findByText("前回の続き m1")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "save" }));
    await waitFor(
      () => {
        expect(saveDraft).toHaveBeenCalledTimes(1);
      },
      { timeout: 2000 },
    );
    expect(saveDraft.mock.calls[0]?.[0]).toMatchObject({
      body: "二文字目",
      channelId: "ch1",
      parentId: "m1",
    });

    await userEvent.click(screen.getByRole("button", { name: "discard" }));
    await waitFor(() => {
      expect(deleteDraft).toHaveBeenCalledWith(
        expect.objectContaining({ channelId: "ch1", parentId: "m1" }),
      );
    });
  });
});
