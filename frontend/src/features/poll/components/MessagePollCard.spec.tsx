import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageSchema, PollMode, PollSchema } from "#/gen/chat/v1/message_pb";
import { PollService } from "#/gen/chat/v1/poll_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { MessagePollCard } from "./MessagePollCard";

import type { VoteRequest } from "#/gen/chat/v1/poll_service_pb";

const poll = create(PollSchema, {
  id: "p1",
  mode: PollMode.TEXT,
  options: [
    { id: "o1", label: "そば", voteCount: 2 },
    { id: "o2", label: "うどん", voteCount: 0 },
  ],
  question: "昼ご飯",
  voterCount: 2,
});

const setup = async (isAuthor: boolean) => {
  const vote = vi.fn<(req: VoteRequest) => void>();
  await renderWithProviders(
    <MessagePollCard poll={poll} isAuthor={isAuthor} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(PollService.method.vote, (req) => {
        vote(req);
        return {
          message: create(MessageSchema, {
            poll: { ...poll, myOptionIds: req.optionIds },
          }),
        };
      });
    },
  );
  return { vote };
};

describe("MessagePollCard", () => {
  test("選択肢と票数を出し、押すと投票して選んだ印を付ける", async () => {
    const { vote } = await setup(false);
    expect(screen.getByText("2 人が投票")).toBeInTheDocument();
    const udon = screen.getByRole("button", { name: /うどん/ });
    expect(udon).toHaveAttribute("aria-pressed", "false");

    await userEvent.click(udon);
    await waitFor(() => {
      expect(vote).toHaveBeenCalledWith(
        expect.objectContaining({ optionIds: ["o2"], pollId: "p1" }),
      );
    });
    await waitFor(() => {
      expect(udon).toHaveAttribute("aria-pressed", "true");
    });
    expect(screen.queryByRole("button", { name: "締め切る" })).not.toBeInTheDocument();
  });

  test("作成者には締め切るボタンを出す", async () => {
    await setup(true);
    expect(screen.getByRole("button", { name: "締め切る" })).toBeInTheDocument();
  });
});
