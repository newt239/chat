import { getRouteApi } from "@tanstack/react-router";

// 右パネルやダイアログの search（workspaceSearchSchema）を子のルートから読む
export const workspaceRoute = getRouteApi("/app/$workspaceId");
