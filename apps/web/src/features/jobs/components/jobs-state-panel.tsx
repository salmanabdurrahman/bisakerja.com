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
    <section
      className="rounded-lg border border-gray-200 bg-gray-50 p-4"
      aria-live="polite"
    >
      <h2 className="text-base font-semibold text-gray-900">{title}</h2>
      <p className="mt-2 text-sm text-gray-600">{description}</p>
      {actionHref && actionLabel ? (
        <a
          href={actionHref}
          className="mt-3 inline-flex rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-white"
        >
          {actionLabel}
        </a>
      ) : null}
    </section>
  );
}
