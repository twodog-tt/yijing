import Link from "next/link";

const FULL_DISCLAIMER =
  "本产品内容仅供娱乐、传统文化学习和自我反思参考，不构成医疗、法律、投资、心理诊断或任何现实决策建议。请结合自身情况理性判断，重要事项请咨询专业人士。";

export default function AboutPage() {
  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 py-8 sm:py-10">
      <Link href="/" className="text-sm text-amber-800 hover:underline">
        ← 返回首页
      </Link>

      <article className="mt-6 space-y-8">
        <header>
          <h1 className="text-2xl font-bold text-stone-900">
            关于 AI 易经问事
          </h1>
          <p className="mt-4 text-base leading-relaxed text-stone-600">
            这是一个基于传统易经卦象的趣味解读与自我反思工具。
          </p>
        </header>

        <section className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-stone-900">起卦方式</h2>
          <p className="mt-3 text-sm leading-relaxed text-stone-600">
            系统会模拟三枚硬币起卦，每一爻可能得到老阴、少阳、少阴、老阳，并据此生成本卦、动爻和变卦。
          </p>
          <ul className="mt-4 space-y-2 text-sm text-stone-600">
            <li>· 6（老阴）：阴爻，动爻</li>
            <li>· 7（少阳）：阳爻</li>
            <li>· 8（少阴）：阴爻</li>
            <li>· 9（老阳）：阳爻，动爻</li>
          </ul>
        </section>

        <section className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-stone-900">内容边界</h2>
          <ul className="mt-4 space-y-2 text-sm leading-relaxed text-stone-600">
            <li>· 不预测未来</li>
            <li>· 不提供医疗建议</li>
            <li>· 不提供法律建议</li>
            <li>· 不提供投资建议</li>
            <li>· 不处理赌博、彩票、伤害、违法相关问题</li>
            <li>· 不提供恐吓式解读</li>
            <li>· 不提供所谓化解、改命、转运服务</li>
          </ul>
        </section>

        <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6">
          <h2 className="text-lg font-semibold text-stone-900">免责声明</h2>
          <p className="mt-3 text-sm leading-relaxed text-stone-700">
            {FULL_DISCLAIMER}
          </p>
        </section>
      </article>
    </main>
  );
}
