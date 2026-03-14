import type { AuthMe } from "@/services/auth";

interface ProfileSummaryProps {
  profile: AuthMe;
}

export function ProfileSummary({ profile }: ProfileSummaryProps) {
  return (
    <section
      className="rounded-lg border border-gray-200 p-4"
      aria-label="Profile summary"
    >
      <h3 className="text-lg font-semibold text-gray-900">Profile</h3>
      <dl className="mt-3 grid gap-2 text-sm text-gray-700">
        <div>
          <dt className="font-medium text-gray-800">Name</dt>
          <dd>{profile.name || "-"}</dd>
        </div>
        <div>
          <dt className="font-medium text-gray-800">Email</dt>
          <dd>{profile.email}</dd>
        </div>
        <div>
          <dt className="font-medium text-gray-800">Role</dt>
          <dd>{profile.role}</dd>
        </div>
      </dl>
    </section>
  );
}
