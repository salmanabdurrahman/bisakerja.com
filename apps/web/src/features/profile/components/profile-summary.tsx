import type { AuthMe } from "@/services/auth";

interface ProfileSummaryProps {
  profile: AuthMe;
}

export function ProfileSummary({ profile }: ProfileSummaryProps) {
  return (
    <section className="bk-card p-6 sm:p-8" aria-label="Profile summary">
      <h3 className="bk-heading-card">Profile</h3>
      <dl className="mt-3 grid gap-3 bk-body text-[#555555]">
        <div>
          <dt className="font-medium text-black">Name</dt>
          <dd>{profile.name || "-"}</dd>
        </div>
        <div>
          <dt className="font-medium text-black">Email</dt>
          <dd>{profile.email}</dd>
        </div>
        <div>
          <dt className="font-medium text-black">Role</dt>
          <dd>{profile.role}</dd>
        </div>
      </dl>
    </section>
  );
}
