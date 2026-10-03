import { IconShieldCheck } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { Tab } from "#/components/ui/Tab/Tab";
import { TabList } from "#/components/ui/TabList/TabList";
import { TabPanel } from "#/components/ui/TabPanel/TabPanel";
import { Tabs } from "#/components/ui/Tabs/Tabs";
import { AdminAppsTab } from "#/features/admin/components/AdminAppsTab";
import { AdminAuditTab } from "#/features/admin/components/AdminAuditTab";
import { AdminInvitationsTab } from "#/features/admin/components/AdminInvitationsTab";
import { AdminMembersTab } from "#/features/admin/components/AdminMembersTab";
import { AdminOverviewTab } from "#/features/admin/components/AdminOverviewTab";
import { AdminPermissionsTab } from "#/features/admin/components/AdminPermissionsTab";
import { useAdminMembers, usePermissions } from "#/features/admin/hooks/useAdminQueries";
import { adminTabValues } from "#/features/admin/schemas";
import { useMyWorkspaceRole } from "#/hooks/useMyWorkspaceRole";

import type { adminSearchSchema } from "#/features/admin/schemas";

import type { z } from "zod";

type AdminTab = z.infer<typeof adminSearchSchema>["tab"];

const adminRoute = getRouteApi("/app/$workspaceId/admin");

export const AdminPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = adminRoute.useParams();
  const { tab } = adminRoute.useSearch();
  const navigate = adminRoute.useNavigate();
  const { data: members, error } = useAdminMembers(workspaceId);
  const { data: permissions } = usePermissions(workspaceId);
  const { data: myRole } = useMyWorkspaceRole(workspaceId);

  const renderTab = (value: AdminTab) => {
    if (error) {
      return (
        <p role="alert" className="m-0 text-caption text-danger">
          {t("admin.loadFailed")}
        </p>
      );
    }
    if (members === undefined) {
      return <Skeleton className="h-64 w-full rounded-xl" />;
    }
    switch (value) {
      case "overview": {
        return <AdminOverviewTab workspaceId={workspaceId} members={members} />;
      }
      case "members": {
        return <AdminMembersTab workspaceId={workspaceId} members={members} />;
      }
      case "invitations": {
        return <AdminInvitationsTab workspaceId={workspaceId} />;
      }
      case "permissions": {
        return permissions === undefined ? (
          <Skeleton className="h-64 w-full rounded-xl" />
        ) : (
          <AdminPermissionsTab
            workspaceId={workspaceId}
            grants={permissions.grants}
            myRole={myRole}
          />
        );
      }
      case "apps": {
        return <AdminAppsTab workspaceId={workspaceId} />;
      }
      case "audit": {
        return <AdminAuditTab workspaceId={workspaceId} members={members} />;
      }
      default: {
        return null;
      }
    }
  };

  return (
    <section className="flex h-full min-h-0 flex-col bg-surface font-sans text-text">
      <PageHeader icon={<IconShieldCheck />} title={t("admin.title")} />
      <Tabs
        selectedKey={tab}
        onSelectionChange={(key) => {
          const next = adminTabValues.find((value) => value === key);
          if (next !== undefined) {
            void navigate({ search: (prev) => ({ ...prev, tab: next }) });
          }
        }}
        className="min-h-0 flex-1"
      >
        <TabList aria-label={t("admin.tabs.label")} className="max-md:px-3">
          {adminTabValues.map((value) => (
            <Tab key={value} id={value}>
              {t(`admin.tabs.${value}`)}
            </Tab>
          ))}
        </TabList>
        {adminTabValues.map((value) => (
          <TabPanel
            key={value}
            id={value}
            className="overflow-y-auto px-4.5 pt-4 pb-6 max-md:px-3.5 max-md:pt-3"
          >
            {renderTab(value)}
          </TabPanel>
        ))}
      </Tabs>
    </section>
  );
};
