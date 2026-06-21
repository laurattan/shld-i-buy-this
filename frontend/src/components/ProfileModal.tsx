import { useEffect, useState } from "react";
import { updateProfile } from "../api";
import type { Profile } from "../types";

interface ProfileModalProps {
  existing: Profile | null;
  onSave: (profile: Profile) => void;
  dismissable: boolean;
  onClose: () => void;
}

export default function ProfileModal({ existing, onSave, dismissable, onClose }: ProfileModalProps) {
  const [name, setName] = useState(existing?.name ?? "");
  const [patterns, setPatterns] = useState(existing?.purchase_patterns ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setName(existing?.name ?? "");
    setPatterns(existing?.purchase_patterns ?? "");
  }, [existing]);

  const handleSave = async () => {
    if (!name.trim()) return;
    setSaving(true);
    setError(null);
    try {
      const profile = await updateProfile({ name: name.trim(), purchase_patterns: patterns.trim() });
      onSave(profile);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save. Try again.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-overlay">
      <button
        type="button"
        aria-label="Close profile"
        className="modal-backdrop"
        onClick={dismissable ? onClose : undefined}
        disabled={!dismissable}
      />

      <div role="dialog" aria-modal="true" aria-labelledby="profile-title" className="modal-card paper-grain shadow-paper">
        <hr className="perf" />
        <div className="modal-body">
          <div className="modal-meta-row">
            <span>customer record</span>
            <span>no. 0001</span>
          </div>

          <h2 id="profile-title" className="modal-title">Your Profile</h2>
          <p className="modal-sub font-accent">so the verdicts actually know you.</p>

          <hr className="perf" style={{ margin: "16px 0" }} />

          <label className="modal-field" style={{ marginBottom: 16 }}>
            <span className="modal-label">Name</span>
            <input
              type="text"
              className="modal-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="what should the receipt call you?"
            />
          </label>

          <label className="modal-field">
            <span className="modal-label">Spending habits</span>
            <textarea
              className="modal-textarea"
              value={patterns}
              onChange={(e) => setPatterns(e.target.value)}
              rows={5}
              placeholder="categories, favorite brands, what you're trying to spend less on…"
            />
          </label>

          <hr className="perf" style={{ margin: "16px 0" }} />

          {error && <p className="modal-error" style={{ marginBottom: 12 }}>{error}</p>}

          <div className="modal-actions">
            {dismissable && (
              <button type="button" className="modal-cancel" onClick={onClose}>
                Cancel
              </button>
            )}
            <button
              type="button"
              className="modal-save"
              onClick={handleSave}
              disabled={!name.trim() || saving}
            >
              {saving ? "Saving…" : "SAVE"}
            </button>
          </div>
        </div>
        <hr className="perf" />
      </div>
    </div>
  );
}
