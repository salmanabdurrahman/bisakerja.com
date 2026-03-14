interface StateCardProps {
  title: string;
  description: string;
}

export function StateCard({ title, description }: StateCardProps) {
  return (
    <section
      aria-label={title}
      className="rounded-lg border border-gray-200 p-4"
    >
      <h2 className="font-medium">{title}</h2>
      <p className="mt-2 text-sm text-gray-600">{description}</p>
    </section>
  );
}
