interface LoadingSpinnerProps {
  label?: string;
}

export default function LoadingSpinner({
  label = "加载中…",
}: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-12 text-stone-500">
      <div
        className="h-8 w-8 animate-spin rounded-full border-2 border-stone-300 border-t-amber-700"
        aria-hidden
      />
      <p className="text-sm">{label}</p>
    </div>
  );
}
