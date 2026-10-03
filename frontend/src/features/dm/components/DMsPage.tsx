import { IconEdit, IconMessageCircle } from "@tabler/icons-react";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { mobileNavTone } from "#/components/block/NavLink/navTone";
import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { cn } from "#/components/ui/styles/styles";
import { DMList } from "#/features/dm/components/DMList";
import { openDialog } from "#/features/layout/utils/overlaySearch";

// モバイルの「DM」タブ
export const DMsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const navigate = useNavigate();

  return (
    <>
      <PageHeader icon={<IconMessageCircle />} title={t("shell.sidebar.dms")}>
        <IconButton
          label={t("shell.sidebar.createDM")}
          onPress={() => {
            void navigate({ search: openDialog({ dialog: "create-dm" }), to: "." });
          }}
        >
          <IconEdit />
        </IconButton>
      </PageHeader>
      <div
        className={cn(
          mobileNavTone,
          "flex min-h-0 flex-1 flex-col gap-px overflow-y-auto p-1.5 text-title font-normal [--nav-row:52px]",
        )}
      >
        <DMList workspaceId={workspaceId} />
      </div>
    </>
  );
};
