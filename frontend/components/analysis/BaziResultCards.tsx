import type { BaziAnalysisView } from "@/lib/bazi";
import { buildElementRows } from "@/lib/bazi";
import SectionReveal from "@/components/motion/SectionReveal";

interface BaziResultCardsProps {
  view: BaziAnalysisView;
  privacyNote?: string;
  revealed?: boolean;
}

export default function BaziResultCards({
  view,
  privacyNote = "出生信息已用于本次简析",
  revealed = true,
}: BaziResultCardsProps) {
  const elementRows = buildElementRows(view.elements);
  const pillars = [
    { label: "年柱", value: view.pillars.year, delay: 240 },
    { label: "月柱", value: view.pillars.month, delay: 320 },
    { label: "日柱", value: view.pillars.day, delay: 400 },
    {
      label: "时柱",
      value: view.hourUnknown ? "时辰未知" : view.pillars.hour,
      delay: 480,
    },
  ];

  return (
    <>
      <SectionReveal active={revealed} delay={0}>
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">方法说明</h2>
          <p className="mt-3 text-sm leading-relaxed text-stone-600">
            {view.methodNote}
          </p>
          <p className="mt-2 text-xs text-stone-400">{privacyNote}</p>
        </section>
      </SectionReveal>

      <SectionReveal active={revealed} delay={80}>
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">简化干支示意</h2>
          <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
            {pillars.map((item) => (
              <SectionReveal key={item.label} active={revealed} delay={item.delay}>
                <div className="pillar-card--flip rounded-xl border border-stone-100 bg-stone-50 px-3 py-3 text-center">
                  <p className="text-xs text-stone-500">{item.label}</p>
                  <p className="mt-1 text-base font-semibold text-stone-900">
                    {item.value || "—"}
                  </p>
                </div>
              </SectionReveal>
            ))}
          </div>
          {view.hourUnknown && (
            <p className="mt-3 text-xs text-stone-500">
              时辰未知，本次不生成时柱
            </p>
          )}
        </section>
      </SectionReveal>

      {(view.baziProfile.elementBalanceType ||
        view.baziProfile.actionStyle ||
        view.baziProfile.reflectionTheme) && (
        <SectionReveal active={revealed} delay={120}>
          <section className="rounded-2xl border border-amber-100 bg-amber-50/60 p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-stone-900">解读视角</h2>
            <ul className="mt-3 space-y-2 text-sm leading-relaxed text-stone-700">
              {view.baziProfile.elementBalanceType && (
                <li>· 五行倾向：{view.baziProfile.elementBalanceType}</li>
              )}
              {view.baziProfile.actionStyle && (
                <li>· 行动风格：{view.baziProfile.actionStyle}</li>
              )}
              {view.baziProfile.reflectionTheme && (
                <li>· 反思主题：{view.baziProfile.reflectionTheme}</li>
              )}
              {view.interpretationLens.pacingHint && (
                <li>· 节奏建议：{view.interpretationLens.pacingHint}</li>
              )}
              {view.baziProfile.dayMasterObservation && (
                <li className="text-stone-600">
                  · {view.baziProfile.dayMasterObservation}
                </li>
              )}
            </ul>
          </section>
        </SectionReveal>
      )}

      <SectionReveal active={revealed} delay={560}>
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">日主</h2>
          <p className="mt-3 text-lg font-semibold text-amber-900">
            {view.dayMaster}
          </p>
        </section>
      </SectionReveal>

      <SectionReveal active={revealed} delay={640}>
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold text-stone-900">五行倾向</h2>
          <div className="mt-4 space-y-2">
            {elementRows.map((row) => (
              <div
                key={row.key}
                className="flex items-center justify-between rounded-lg bg-stone-50 px-3 py-2"
              >
                <span
                  className={`text-sm text-stone-600 element-row-glow element-row-glow--${row.key}`}
                >
                  {row.label}
                </span>
                <span className="text-sm font-medium text-stone-900">
                  {row.value}
                </span>
              </div>
            ))}
          </div>
        </section>
      </SectionReveal>

      {view.reflectionFocus && (
        <SectionReveal active={revealed} delay={720}>
          <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-stone-900">反思焦点</h2>
            <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-stone-700">
              {view.reflectionFocus}
            </p>
          </section>
        </SectionReveal>
      )}

      {view.actionSuggestions.length > 0 && (
        <SectionReveal active={revealed} delay={800}>
          <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-stone-900">行动建议</h2>
            <ul className="mt-3 space-y-2 text-sm leading-relaxed text-stone-700">
              {view.actionSuggestions.map((item, index) => (
                <li key={`${index}-${item.slice(0, 16)}`}>· {item}</li>
              ))}
            </ul>
          </section>
        </SectionReveal>
      )}
    </>
  );
}
