import type { Hexagram } from "@/lib/types";

interface HexagramCardProps {
  title: string;
  hexagram: Hexagram;
  accent?: "primary" | "changed";
}

export default function HexagramCard({
  title,
  hexagram,
  accent = "primary",
}: HexagramCardProps) {
  const borderClass =
    accent === "primary"
      ? "border-amber-200 bg-gradient-to-b from-amber-50/80 to-white"
      : "border-stone-200 bg-gradient-to-b from-stone-50 to-white";

  return (
    <div className={`rounded-2xl border p-5 shadow-sm ${borderClass}`}>
      <p className="text-xs font-medium tracking-widest text-amber-800">
        {title}
      </p>
      <p className="mt-3 text-2xl font-bold text-stone-900">
        {hexagram.full_name}
      </p>
      <p className="mt-1 text-sm text-stone-600">
        第 {hexagram.number} 卦 · {hexagram.name}
      </p>
      <p className="mt-3 font-mono text-xs tracking-wider text-stone-400">
        {hexagram.binary_code}
      </p>
      {hexagram.summary && (
        <p className="mt-4 border-t border-stone-100 pt-3 text-sm leading-relaxed text-stone-600">
          {hexagram.summary}
        </p>
      )}
    </div>
  );
}
