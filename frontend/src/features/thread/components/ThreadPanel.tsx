import { useCallback, useEffect, useRef } from "react";

import { Divider, Loader, Stack, Text } from "@mantine/core";
import { useAtom, useAtomValue, useSetAtom } from "jotai";

import { MessageItem } from "#/features/message/components/MessageItem";
import { ThreadReplyInput } from "#/features/message/components/ThreadReplyInput";
import { ThreadReplyList } from "#/features/message/components/ThreadReplyList";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useThreadReplies, useSendThreadReply } from "#/features/message/hooks/useThread";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "#/providers/store/workspace";

type ThreadPanelProps = {
  threadId: string;
};

export const ThreadPanel = ({ threadId }: ThreadPanelProps) => {
  const currentUser = useAtomValue(userAtom);
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const { data: threadData, isLoading, isError, error } = useThreadReplies(threadId);
  const sendReply = useSendThreadReply();
  const repliesEndRef = useRef<HTMLDivElement>(null);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  const scrollToBottom = () => {
    repliesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (!isLoading) {
      scrollToBottom();
    }
  }, [threadData, isLoading]);

  useEffect(() => {
    if (sendReply.isSuccess) {
      scrollToBottom();
    }
  }, [sendReply.isSuccess]);

  const handleCopyLink = useCopyMessageLink(currentWorkspaceId, currentChannelId);

  const handleCreateThread = useCallback(
    (msgId: string) => {
      setRightSidePanelView({ threadId: msgId, type: "thread" });
    },
    [setRightSidePanelView],
  );

  const handleSendReply = useCallback(
    (body: string, attachmentIds: string[]) => {
      if (currentChannelId !== null) {
        sendReply.mutate({ attachmentIds, body, channelId: currentChannelId, parentId: threadId });
      }
    },
    [sendReply, currentChannelId, threadId],
  );

  if (!currentWorkspaceId || !currentChannelId) {
    return (
      <div className="p-4">
        <Text c="dimmed" size="sm">
          ワークスペースまたはチャンネルが選択されていません
        </Text>
      </div>
    );
  }

  return (
    <div>
      <div className="flex h-full flex-col">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : isError ? (
          <Text c="red" size="sm">
            {error.message}
          </Text>
        ) : threadData?.parentMessage ? (
          <>
            <div className="flex-1 overflow-y-auto">
              <Stack gap="md">
                {/* 親メッセージ */}
                <MessageItem
                  message={threadData.parentMessage}
                  currentUserId={currentUser?.id ?? null}
                  onCopyLink={handleCopyLink}
                  onCreateThread={handleCreateThread}
                />

                <Divider label={`${threadData.replies.length}件の返信`} labelPosition="center" />

                {/* 返信一覧 */}
                <ThreadReplyList
                  replies={threadData.replies}
                  currentUserId={currentUser?.id ?? null}
                  workspaceId={currentWorkspaceId}
                  channelId={currentChannelId}
                />

                <div ref={repliesEndRef} />
              </Stack>
            </div>

            {/* 返信入力 */}
            <div className="shrink-0">
              <ThreadReplyInput
                channelId={currentChannelId}
                onSubmit={handleSendReply}
                isPending={sendReply.isPending}
                isError={sendReply.isError}
                errorMessage={sendReply.error?.message}
              />
            </div>
          </>
        ) : (
          <Text c="dimmed" size="sm">
            スレッドが見つかりません
          </Text>
        )}
      </div>
    </div>
  );
};
