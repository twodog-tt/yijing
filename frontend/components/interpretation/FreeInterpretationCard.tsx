import { sanitizeInternalTerms } from "@/lib/display-text";

const FOOTER_DISCLAIMER =
  "本内容仅供娱乐和传统文化参考，不构成现实决策建议。";

interface FreeInterpretationCardProps {
  content: string;
}

export default function FreeInterpretationCard({
  content,
}: FreeInterpretationCardProps) {
  const safeContent = sanitizeInternalTerms(content);

  return (
    <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:p-7">
      <h2 className="text-lg font-bold text-stone-900">免费解读</h2>
      <div className="mt-5 text-sm leading-7 text-stone-700 sm:leading-8">
        <p className="whitespace-pre-wrap">{safeContent || "暂无免费解读"}</p>
      </div>
      <p className="mt-6 border-t border-stone-100 pt-4 text-xs leading-relaxed text-stone-400">
        {FOOTER_DISCLAIMER}
      </p>
    </section>
  );
}
