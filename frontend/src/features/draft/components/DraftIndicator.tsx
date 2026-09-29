import { IconPencil } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { useHasDraft } from "../hooks/useDrafts";

type DraftIndicatorProps = {
  workspaceId: string;
  channelId: string;
};

// サイドバーのチャンネルに、書きかけの下書きがあることを示す
export const DraftIndicator = ({ workspaceId, channelId }: DraftIndicatorProps) => {
  const { t } = useTranslation();
  const hasDraft = useHasDraft(workspaceId, channelId);
  return hasDraft ? <IconPencil aria-label={t("draft.sidebar.hasDraft")} role="img" /> : null;
};
