import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useChannelMemberActions } from "../hooks/useChannelMemberActions";

type JoinChannelBarProps = {
  workspaceId: string;
  channelId: string;
  channelName: string;
};

// 未参加のチャンネルをプレビューしているとき、入力欄の代わりに出す
export const JoinChannelBar = ({ workspaceId, channelId, channelName }: JoinChannelBarProps) => {
  const { t } = useTranslation();
  const { join } = useChannelMemberActions(workspaceId);

  return (
    <div className="shrink-0 px-4.5 pb-3 font-sans max-md:px-2.5 max-md:pb-2">
      <div className="flex flex-col items-center gap-2 rounded-lg border border-border bg-sunken px-4 py-3 text-center">
        <p className="m-0 text-body-sm font-semibold">
          {t("channel.preview.notice", { name: channelName })}
        </p>
        <p className="m-0 text-caption text-muted">{t("channel.preview.notJoined")}</p>
        <Button
          isPending={join.isPending}
          onPress={() => {
            join.mutate(
              { channelId },
              {
                onSuccess: () => {
                  toast(t("channel.browse.joinedToast", { name: channelName }), {
                    tone: "success",
                  });
                },
              },
            );
          }}
        >
          {t("channel.preview.join")}
        </Button>
      </div>
    </div>
  );
};
