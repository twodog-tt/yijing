import type { FullReport } from "@/lib/types";
import { parseFullReport } from "@/lib/api";

interface FullInterpretationProps {
  content: FullReport | string;
}

export default function FullInterpretation({
  content,
}: FullInterpretationProps) {
  const report = parseFullReport(content);

  if (!report) {
    return (
      <section className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
        <h2 className="text-lg font-bold text-stone-900">完整解读</h2>
        <p className="mt-4 whitespace-pre-wrap text-sm leading-relaxed text-stone-700">
          {typeof content === "string" ? content : ""}
        </p>
      </section>
    );
  }

  return (
    <section className="space-y-4">
      <div className="rounded-2xl border border-amber-200 bg-amber-50 p-5 shadow-sm">
        <p className="text-xs tracking-widest text-amber-800">一句话总结</p>
        <p className="mt-2 text-base font-semibold leading-relaxed text-stone-900">
          {report.summary}
        </p>
      </div>

      <TextCard title="总体判断" content={report.overall} />
      <TextCard title="当前处境" content={report.current_state} />
      <TextCard title="机会点" content={report.opportunity} />
      <TextCard title="风险点" content={report.risk} />

      <ListCard title="行动建议" items={report.action_steps} ordered />
      <TextCard title="情绪提醒" content={report.emotion_reminder} />
      <ListCard
        title="自我反思问题"
        items={report.reflection_questions}
      />

      <p className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 text-xs leading-relaxed text-stone-500">
        {report.disclaimer}
      </p>
    </section>
  );
}

function TextCard({ title, content }: { title: string; content: string }) {
  return (
    <div className="rounded-xl border border-stone-200 bg-white p-5 shadow-sm">
      <h3 className="text-sm font-semibold text-stone-800">{title}</h3>
      <p className="mt-2 text-sm leading-relaxed text-stone-700">{content}</p>
    </div>
  );
}

function ListCard({
  title,
  items,
  ordered = false,
}: {
  title: string;
  items: string[];
  ordered?: boolean;
}) {
  const Tag = ordered ? "ol" : "ul";
  return (
    <div className="rounded-xl border border-stone-200 bg-white p-5 shadow-sm">
      <h3 className="text-sm font-semibold text-stone-800">{title}</h3>
      <Tag
        className={`mt-3 space-y-2 text-sm leading-relaxed text-stone-700 ${ordered ? "list-decimal pl-5" : "list-disc pl-5"}`}
      >
        {items.map((item, i) => (
          <li key={i}>{item}</li>
        ))}
      </Tag>
    </div>
  );
}
