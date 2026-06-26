const DISCLAIMER =
  "仅供娱乐、传统文化学习和自我反思参考，不构成预测、医疗、法律、投资或重大决策建议。";

export default function DisclaimerBanner() {
  return (
    <div
      role="note"
      className="border-b border-amber-200/60 bg-amber-50 px-4 py-2 text-center text-xs leading-relaxed text-amber-900 sm:text-sm"
    >
      {DISCLAIMER}
    </div>
  );
}

export { DISCLAIMER };
