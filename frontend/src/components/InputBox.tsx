import { useState } from "react";

interface InputBoxProps {
  onSubmit: (product: string) => void;
  loading: boolean;
  needsProfile?: boolean;
  onRequestProfile?: () => void;
}

export default function InputBox({
  onSubmit,
  loading,
  needsProfile = false,
  onRequestProfile,
}: InputBoxProps) {
  const [value, setValue] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (needsProfile) {
      onRequestProfile?.();
      return;
    }
    if (!value.trim() || loading) return;
    onSubmit(value.trim());
    setValue("");
  };

  return (
    <div className="input-bar">
      <div className="input-body">
        <form className="input-form" onSubmit={handleSubmit}>
          <input
            type="text"
            className="input-field"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={
              needsProfile ? "set up your profile first…" : "what are you thinking of buying?"
            }
            aria-label="Product to evaluate"
            />
          <button
            type="submit"
            className="input-submit"
            disabled={loading}
          >
            {needsProfile ? "Set up" : "Print"}
          </button>
        </form>
      </div>
    </div>
  );
}
