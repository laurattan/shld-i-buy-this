import type { CSSProperties } from "react";
import { VERDICT_META, VERDICT_ROTATIONS, type Verdict } from "../types";

interface VerdictStampProps {
  verdict: Verdict;
  animate?: boolean;
}

export default function VerdictStamp({ verdict, animate = false }: VerdictStampProps) {
  const meta = VERDICT_META[verdict];
  const rot = VERDICT_ROTATIONS[verdict];
  const style: CSSProperties = {
    color: meta.color,
    borderColor: meta.color,
    boxShadow: `0 0 0 2px ${meta.color}`,
    ["--stamp-rot" as string]: `${rot}deg`,
    transform: `rotate(${rot}deg)`,
  };
  return (
    <span
      role="img"
      aria-label={`Verdict: ${meta.label}`}
      className={`verdict-stamp ink-stamp${animate ? " animate-stamp" : ""}`}
      style={style}
    >
      {meta.label}
    </span>
  );
}
