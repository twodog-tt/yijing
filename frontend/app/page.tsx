import Link from "next/link";

export default function Home() {
  return (
    <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 py-10 sm:py-16">
      <section className="rounded-2xl border border-stone-200 bg-white p-8 shadow-sm sm:p-10">
        <p className="text-sm tracking-[0.2em] text-amber-800">易经问事</p>
        <h1 className="mt-4 text-3xl font-bold leading-tight text-stone-900 sm:text-4xl">
          心中有事，起一卦。
        </h1>
        <p className="mt-5 text-base leading-relaxed text-stone-600">
          用传统易经卦象，帮你整理当下处境、风险与下一步行动。
        </p>

        <div className="mt-8 flex flex-col gap-3">
          <Link
            href="/ask"
            className="inline-flex items-center justify-center rounded-xl bg-stone-900 px-6 py-3.5 text-sm font-semibold text-white transition hover:bg-stone-800"
          >
            立即起卦
          </Link>
          <Link
            href="/today"
            className="inline-flex items-center justify-center rounded-xl border border-amber-300 bg-amber-50 px-6 py-3.5 text-sm font-semibold text-amber-900 transition hover:bg-amber-100"
          >
            查看今日一卦
          </Link>
          <div className="flex gap-3">
            <Link
              href="/history"
              className="inline-flex flex-1 items-center justify-center rounded-xl border border-stone-300 bg-white px-6 py-3.5 text-sm font-semibold text-stone-800 transition hover:bg-stone-50"
            >
              历史记录
            </Link>
            <Link
              href="/about"
              className="inline-flex flex-1 items-center justify-center rounded-xl border border-stone-300 bg-white px-6 py-3.5 text-sm font-semibold text-stone-800 transition hover:bg-stone-50"
            >
              关于说明
            </Link>
          </div>
        </div>
      </section>

      <section className="mt-6 rounded-xl border border-amber-200/80 bg-gradient-to-r from-amber-50 to-stone-50 p-5">
        <p className="text-sm font-medium text-amber-900">今日运势</p>
        <p className="mt-1 text-xs leading-relaxed text-stone-600">
          用一卦整理今天的状态、节奏与行动提醒。
        </p>
        <Link
          href="/today"
          className="mt-3 inline-flex text-sm font-semibold text-amber-800 hover:underline"
        >
          查看今日一卦 →
        </Link>
      </section>

      <section className="mt-6 rounded-xl border border-stone-200 bg-stone-50 p-5 text-xs leading-relaxed text-stone-500">
        仅供娱乐和传统文化参考，不构成任何预测、医疗、法律或投资建议。
        <Link href="/about" className="ml-1 text-amber-800 hover:underline">
          查看完整说明
        </Link>
      </section>
    </main>
  );
}
