import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";

import { useDMs } from "../hooks/useDM";
import { DMRow } from "./DMRow";

type DMListProps = {
  workspaceId: string;
};

export const DMList = ({ workspaceId }: DMListProps) => {
  const { t } = useTranslation();
  const { data: dms, isLoading } = useDMs(workspaceId);

  if (isLoading) {
    return <Skeleton className="mx-2 my-1 h-4 w-28 bg-(--nav-hover)" />;
  }

  if (!dms || dms.length === 0) {
    return (
      <p className="m-0 px-2 py-1 text-caption text-(--nav-muted)">{t("shell.sidebar.noDMs")}</p>
    );
  }

  return dms.map((dm) => <DMRow key={dm.id} workspaceId={workspaceId} dm={dm} />);
};
