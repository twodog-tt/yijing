import { sanitizeInternalTerms } from "@/lib/display-text";

interface AnalysisFullContentProps {
  content: string;
  revealed?: boolean;
}

export default function AnalysisFullContent({
  content,
  revealed = true,
}: AnalysisFullContentProps) {
  const safeContent = sanitizeInternalTerms(content);

  return (
    <section
      id="full-report"
      className={`scroll-mt-4 rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:p-7 ${
        revealed ? "full-fade-in" : "opacity-0"
      }`}
    >
      <h2 className="text-lg font-bold text-stone-900">完整报告</h2>
      <div className="mt-5 text-sm leading-7 text-stone-700 sm:leading-8">
        <p className="whitespace-pre-wrap">{safeContent}</p>
      </div>
    </section>
  );
}
