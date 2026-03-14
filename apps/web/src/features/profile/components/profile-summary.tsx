import type { AuthMe } from "@/services/auth";

interface ProfileSummaryProps {
  profile: AuthMe;
}

export function ProfileSummary({ profile }: ProfileSummaryProps) {
  return (
    <section className="bk-card p-6 sm:p-8" aria-label="Profile summary">
      <h3 className="text-lg font-semibold text-slate-900">Profile</h3>
      <dl className="mt-3 grid gap-3 text-sm text-slate-700">
        <div>
          <dt className="font-medium text-slate-800">Name</dt>
          <dd>{profile.name || "-"}</dd>
        </div>
        <div>
          <dt className="font-medium text-slate-800">Email</dt>
          <dd>{profile.email}</dd>
        </div>
        <div>
          <dt className="font-medium text-slate-800">Role</dt>
          <dd>{profile.role}</dd>
        </div>
      </dl>
    </section>
  );
}
