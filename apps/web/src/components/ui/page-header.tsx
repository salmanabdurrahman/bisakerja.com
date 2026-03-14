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
      {eyebrow ? (
        <p className="text-[14px] font-medium uppercase tracking-[0.18em] text-[#888888]">
          {eyebrow}
        </p>
      ) : null}
      <div className="flex flex-col lg:flex-row lg:items-end justify-between gap-8">
        <div className="max-w-3xl grid gap-4">
          <h2 className="text-[40px] sm:text-[48px] font-normal tracking-tight text-black font-display leading-[1.1]">
            {title}
          </h2>
          {description ? (
            <p className="text-[16px] sm:text-[18px] font-normal text-[#666666] leading-relaxed">
              {description}
            </p>
          ) : null}
        </div>
        {actions ? (
          <div className="flex flex-wrap gap-3 mt-4 lg:mt-0">{actions}</div>
        ) : null}
      </div>
    </header>
  );
}
