import type { QimenAnalysisView } from "@/lib/qimen";

interface QimenResultCardsProps {
  view: QimenAnalysisView;
  privacyNote?: string;
}

export default function QimenResultCards({
  view,
  privacyNote = "问事主题已用于本次局势梳理",
}: QimenResultCardsProps) {
  return (
    <>
      <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
        <h2 className="text-sm font-semibold text-stone-900">方法说明</h2>
        <p className="mt-3 text-sm leading-relaxed text-stone-600">
          {view.methodNote}
        </p>
        <p className="mt-2 text-xs text-stone-400">{privacyNote}</p>
      </section>

      <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
        <h2 className="text-sm font-semibold text-stone-900">局势梳理</h2>
        <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-stone-700">
          {view.situationOverview}
        </p>
      </section>

      {view.riskObservations.length > 0 && (
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">风险观察</h2>
          <ul className="mt-3 space-y-2 text-sm leading-relaxed text-stone-700">
            {view.riskObservations.map((item, index) => (
              <li key={`risk-${index}`}>· {item}</li>
            ))}
          </ul>
        </section>
      )}

      {view.actionPacing && (
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">行动节奏</h2>
          <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-stone-700">
            {view.actionPacing}
          </p>
        </section>
      )}

      {view.reflectionQuestions.length > 0 && (
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">自我反思问题</h2>
          <ul className="mt-3 space-y-2 text-sm leading-relaxed text-stone-700">
            {view.reflectionQuestions.map((item, index) => (
              <li key={`reflection-${index}`}>· {item}</li>
            ))}
          </ul>
        </section>
      )}

      {view.actionSuggestions.length > 0 && (
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">行动建议</h2>
          <ul className="mt-3 space-y-2 text-sm leading-relaxed text-stone-700">
            {view.actionSuggestions.map((item, index) => (
              <li key={`action-${index}`}>· {item}</li>
            ))}
          </ul>
        </section>
      )}

      {view.limits.length > 0 && (
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">简化说明</h2>
          <ul className="mt-3 space-y-2 text-sm leading-relaxed text-stone-600">
            {view.limits.map((item, index) => (
              <li key={`limit-${index}`}>· {item}</li>
            ))}
          </ul>
        </section>
      )}
    </>
  );
}
