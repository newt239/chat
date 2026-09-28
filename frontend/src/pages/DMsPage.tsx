import { useState } from "react";

import { IconEdit, IconMessageCircle } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton";
import { cn } from "#/components/ui/styles";
import { CreateDMModal } from "#/features/dm/components/CreateDMModal";
import { DMList } from "#/features/dm/components/DMList";
import { PageHeader } from "#/features/layout/components/PageHeader";
import { mobileNavTone } from "#/features/layout/utils/navTone";

// モバイルの「DM」タブ
export const DMsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const [isCreating, setIsCreating] = useState(false);

  return (
    <>
      <PageHeader icon={<IconMessageCircle />} title={t("shell.sidebar.dms")}>
        <IconButton
          label={t("shell.sidebar.createDM")}
          onPress={() => {
            setIsCreating(true);
          }}
        >
          <IconEdit />
        </IconButton>
      </PageHeader>
      <div
        className={cn(
          mobileNavTone,
          "flex min-h-0 flex-1 flex-col gap-px overflow-y-auto p-1.5 text-[15px] [--nav-row:52px]",
        )}
      >
        <DMList workspaceId={workspaceId} />
      </div>
      <CreateDMModal
        workspaceId={workspaceId}
        opened={isCreating}
        onClose={() => {
          setIsCreating(false);
        }}
      />
    </>
  );
};
