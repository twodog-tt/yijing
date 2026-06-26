import type { CSSProperties, ReactNode } from "react";

interface SectionRevealProps {
  active?: boolean;
  delay?: number;
  staticSection?: boolean;
  children: ReactNode;
  className?: string;
}

export default function SectionReveal({
  active = true,
  delay = 0,
  staticSection = false,
  children,
  className = "",
}: SectionRevealProps) {
  const style: CSSProperties | undefined =
    active && !staticSection ? { animationDelay: `${delay}ms` } : undefined;

  return (
    <div
      className={`section-reveal ${
        staticSection ? "section-reveal--static" : ""
      } ${active && !staticSection ? "section-reveal--active" : ""} ${className}`.trim()}
      style={style}
    >
      {children}
    </div>
  );
}
