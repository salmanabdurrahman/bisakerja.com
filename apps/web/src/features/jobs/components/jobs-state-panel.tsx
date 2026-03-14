import { ButtonLink } from "@/components/ui/button";

interface JobsStatePanelProps {
  title: string;
  description: string;
  actionHref?: string;
  actionLabel?: string;
}

export function JobsStatePanel({
  title,
  description,
  actionHref,
  actionLabel,
}: JobsStatePanelProps) {
  return (
    <section className="bk-card-muted p-5" aria-live="polite">
      <h2 className="text-base font-semibold text-slate-900">{title}</h2>
      <p className="mt-2 text-sm text-slate-600">{description}</p>
      {actionHref && actionLabel ? (
        <ButtonLink
          href={actionHref}
          variant="outline"
          size="sm"
          className="mt-3"
        >
          {actionLabel}
        </ButtonLink>
      ) : null}
    </section>
  );
}
