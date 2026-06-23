import type { Line } from "@/lib/types";

interface LineChartProps {
  lines: Line[];
  title?: string;
}

function YangLine({ moving }: { moving: boolean }) {
  return (
    <div className="flex min-h-[28px] flex-1 items-center gap-2">
      <div
        className={`h-2 w-full rounded-sm shadow-sm ${
          moving ? "bg-amber-500 ring-2 ring-amber-200" : "bg-stone-800"
        }`}
      />
      {moving && <MovingBadge />}
    </div>
  );
}

function YinLine({ moving }: { moving: boolean }) {
  return (
    <div className="flex min-h-[28px] flex-1 items-center gap-2">
      <div className="flex w-full gap-2 sm:gap-3">
        <div
          className={`h-2 flex-1 rounded-sm shadow-sm ${
            moving ? "bg-amber-500 ring-2 ring-amber-200" : "bg-stone-800"
          }`}
        />
        <div
          className={`h-2 flex-1 rounded-sm shadow-sm ${
            moving ? "bg-amber-500 ring-2 ring-amber-200" : "bg-stone-800"
          }`}
        />
      </div>
      {moving && <MovingBadge />}
    </div>
  );
}

function MovingBadge() {
  return (
    <span className="shrink-0 rounded-md bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-800">
      动
    </span>
  );
}

export default function LineChart({ lines, title }: LineChartProps) {
  const sorted = [...lines].sort((a, b) => b.position - a.position);

  return (
    <div className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
      {title && (
        <p className="mb-5 text-sm font-semibold text-stone-800">{title}</p>
      )}
      <div className="flex flex-col gap-4">
        {sorted.map((line) => (
          <div
            key={line.position}
            className={`flex items-center gap-3 rounded-lg px-2 py-1 ${
              line.is_moving === 1 ? "bg-amber-50/80" : ""
            }`}
          >
            <span className="w-10 shrink-0 text-right text-xs font-medium text-stone-500">
              {line.position}爻
            </span>
            {line.is_yang === 1 ? (
              <YangLine moving={line.is_moving === 1} />
            ) : (
              <YinLine moving={line.is_moving === 1} />
            )}
          </div>
        ))}
      </div>
      <p className="mt-4 text-center text-xs text-stone-400">上 ↑　下 ↓</p>
    </div>
  );
}
