interface StateCardProps {
  title: string;
  description: string;
}

export function StateCard({ title, description }: StateCardProps) {
  return (
    <section aria-label={title} className="bk-card grid gap-3 p-6 sm:p-8">
      <h2 className="text-lg font-semibold tracking-tight text-slate-900">
        {title}
      </h2>
      <p className="bk-subtle">{description}</p>
    </section>
  );
}
