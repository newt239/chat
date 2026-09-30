import type { ReactNode } from "react";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { useIsMobile } from "#/hooks/useMediaQuery";

type SettingsLayoutProps = {
  icon: ReactNode;
  title: string;
  sectionTitle: string;
  // SettingsNavLink を並べる。モバイルでは項目ごとに積んだ画面なので出さない
  nav: ReactNode;
  children: ReactNode;
};

export const SettingsLayout = ({
  icon,
  title,
  sectionTitle,
  nav,
  children,
}: SettingsLayoutProps) => {
  const isMobile = useIsMobile();

  return (
    <section className="flex h-full min-h-0 flex-col bg-surface font-sans text-text">
      <PageHeader icon={icon} title={isMobile ? sectionTitle : title} />
      <div className="flex min-h-0 flex-1 overflow-y-auto">
        <div className="mx-auto flex w-full max-w-4xl gap-8 px-[18px] pt-5 pb-8 max-md:px-3.5 max-md:pt-3">
          {!isMobile && (
            <nav
              aria-label={title}
              className="sticky top-0 flex w-48 shrink-0 flex-col gap-px self-start"
            >
              {nav}
            </nav>
          )}
          <div className="flex min-w-0 flex-1 flex-col gap-3">
            {!isMobile && <h2 className="m-0 text-title">{sectionTitle}</h2>}
            {children}
          </div>
        </div>
      </div>
    </section>
  );
};
