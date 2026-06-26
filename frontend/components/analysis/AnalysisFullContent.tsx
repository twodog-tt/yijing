interface AnalysisFullContentProps {
  content: string;
}

export default function AnalysisFullContent({
  content,
}: AnalysisFullContentProps) {
  return (
    <section
      id="full-report"
      className="scroll-mt-4 rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:p-7"
    >
      <h2 className="text-lg font-bold text-stone-900">完整报告</h2>
      <div className="mt-5 text-sm leading-7 text-stone-700 sm:leading-8">
        <p className="whitespace-pre-wrap">{content}</p>
      </div>
    </section>
  );
}
