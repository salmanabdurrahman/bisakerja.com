import type { ReactNode } from "react";

import { cn } from "@/lib/utils/cn";

interface PageHeaderProps {
  title: string;
  description?: string;
  eyebrow?: string;
  actions?: ReactNode;
  className?: string;
}

/**
 * PageHeader renders a consistent page hero block.
 */
export function PageHeader({
  title,
  description,
  eyebrow,
  actions,
  className,
}: PageHeaderProps) {
  return (
    <header className={cn("grid gap-5 mb-12", className)}>
      {eyebrow ? <p className="bk-eyebrow">{eyebrow}</p> : null}
      <div className="flex flex-col lg:flex-row lg:items-end justify-between gap-8">
        <div className="max-w-3xl grid gap-4">
          <h2 className="bk-heading-page">{title}</h2>
          {description ? <p className="bk-body-lg">{description}</p> : null}
        </div>
        {actions ? (
          <div className="flex flex-wrap gap-3 mt-4 lg:mt-0">{actions}</div>
        ) : null}
      </div>
    </header>
  );
}
