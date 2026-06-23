const DISCLAIMER =
  "仅供娱乐和传统文化参考，不构成任何预测、医疗、法律或投资建议。";

export default function DisclaimerBanner() {
  return (
    <div
      role="note"
      className="border-b border-amber-200/60 bg-amber-50 px-4 py-2 text-center text-sm text-amber-900"
    >
      {DISCLAIMER}
    </div>
  );
}

export { DISCLAIMER };
