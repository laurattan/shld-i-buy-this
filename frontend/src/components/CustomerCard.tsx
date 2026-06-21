import type { Profile } from "../types";

interface CustomerCardProps {
  profile: Profile | null;
  onEditProfile: () => void;
}

export default function CustomerCard({ profile, onEditProfile }: CustomerCardProps) {
  const hasProfile = !!profile?.name;

  return (
    <div className="left-panel">
      <div className="customer-card-wrap">
        <button
          type="button"
          className="customer-card paper-grain shadow-paper"
          onClick={onEditProfile}
          aria-label="Edit profile"
        >
          <div className="customer-card-holes" aria-hidden="true">
            {Array.from({ length: 7 }).map((_, i) => (
              <span key={i} className="customer-card-hole" />
            ))}
          </div>

          <div className="customer-card-label-row">
            <span>customer card</span>
            <span className="customer-card-edit-hint">edit ▸</span>
          </div>

          <p className="customer-card-name">
            {hasProfile ? profile!.name : "no name set"}
          </p>

          <p className="customer-card-habits font-accent">
            {profile?.purchase_patterns?.trim() || "tap to add your spending habits."}
          </p>
        </button>
      </div>
    </div>
  );
}
