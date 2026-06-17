import { useState } from "react";
import { updateProfile } from "../api";
import type { Profile } from "../types";

interface ProfileModalProps {
  existing: Profile | null;
  onSave: (profile: Profile) => void;
  dismissable: boolean;
}

export default function ProfileModal({ existing, onSave, dismissable }: ProfileModalProps) {
  const [name, setName] = useState(existing?.name ?? "");
  const [patterns, setPatterns] = useState(existing?.purchase_patterns ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
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
      <div className="modal">
        {dismissable && (
          <button type="button" className="modal-close" onClick={() => onSave(existing!)}>
            ✕
          </button>
        )}
        <p>
          {existing
            ? "Edit your profile."
            : "Welcome to shld i buy this. Let's set up your profile."}
        </p>
        <label>
          <span>Name</span>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="what should the receipt call you?"
          />
        </label>
        <p>
          For each line, describe your top spending categories, what you own, and your go-to brands.
          <br />
          Example: "electronics: AirPods, MacBook, Sony headphones, brands: Apple, Sony. clothing: casual basics, brands: Uniqlo, Zara."
        </p>
        <textarea
          value={patterns}
          onChange={(e) => setPatterns(e.target.value)}
          rows={6}
        />
        {error && <p style={{ color: "red", margin: 0 }}>{error}</p>}
        <button
          type="button"
          onClick={handleSave}
          disabled={name.trim() === "" || saving}
        >
          {saving ? "Saving…" : "Save"}
        </button>
      </div>
    </div>
  );
}
