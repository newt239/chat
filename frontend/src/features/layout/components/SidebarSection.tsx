import type { ReactNode } from "react";

import { IconChevronDown, IconPlus } from "@tabler/icons-react";
import { useAtom } from "jotai";
import { AnimatePresence, motion } from "motion/react";
import { Button } from "react-aria-components";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { focusRing } from "#/components/ui/styles/styles";
import { collapsedSidebarSectionsAtom } from "#/features/layout/atoms";
import { transitions } from "#/lib/motion";

type SidebarSectionProps = {
  id: string;
  title: string;
  onAdd: { label: string; onPress: () => void } | null;
  menu: ReactNode;
  children: ReactNode;
};

export const SidebarSection = ({ id, title, onAdd, menu, children }: SidebarSectionProps) => {
  const [collapsed, setCollapsed] = useAtom(collapsedSidebarSectionsAtom);
  const isCollapsed = collapsed[id] ?? false;
  const handleAdd = onAdd?.onPress;
  const handleToggle = () => {
    setCollapsed({ ...collapsed, [id]: !isCollapsed });
  };

  return (
    <section className="flex flex-col">
      <div className="flex items-center gap-1 pt-3 pr-2 pb-0.5 pl-2.5">
        <Button
          aria-expanded={!isCollapsed}
          onPress={handleToggle}
          className={`flex min-w-0 flex-1 cursor-pointer items-center gap-1 rounded-sm text-left text-xs font-semibold text-(--nav-muted) data-hovered:text-(--nav-strong) ${focusRing}`}
        >
          <IconChevronDown
            aria-hidden
            className={`size-3 shrink-0 transition-transform motion-reduce:transition-none ${isCollapsed ? "-rotate-90" : ""}`}
          />
          <span className="truncate">{title}</span>
        </Button>
        {menu}
        {onAdd && (
          <IconButton
            label={onAdd.label}
            onPress={handleAdd}
            className="size-6 text-(--nav-muted) data-hovered:bg-(--nav-hover) data-hovered:text-(--nav-strong) [&_svg]:size-3.5"
          >
            <IconPlus />
          </IconButton>
        )}
      </div>
      <AnimatePresence initial={false}>
        {!isCollapsed && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={transitions.base}
            className="flex flex-col gap-px overflow-hidden px-1.5"
          >
            {children}
          </motion.div>
        )}
      </AnimatePresence>
    </section>
  );
};
