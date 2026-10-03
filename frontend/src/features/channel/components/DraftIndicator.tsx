import { useQuery } from "@connectrpc/connect-query";
import { IconPencil } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { DraftService } from "#/gen/chat/v1/draft_service_pb";

type DraftIndicatorProps = {
  workspaceId: string;
  channelId: string;
};

// サイドバーのチャンネルに、書きかけの下書きがあることを示す。スレッドの下書きも数える
export const DraftIndicator = ({ workspaceId, channelId }: DraftIndicatorProps) => {
  const { t } = useTranslation();
  const { data: hasDraft } = useQuery(
    DraftService.method.listDrafts,
    { workspaceId },
    { select: (res) => res.drafts.some((draft) => draft.channelId === channelId) },
  );
  return hasDraft ? <IconPencil aria-label={t("draft.sidebar.hasDraft")} role="img" /> : null;
};
