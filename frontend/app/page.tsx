import ElementOrbit from "@/components/motion/ElementOrbit";
import QimenScanGrid from "@/components/motion/QimenScanGrid";
import Link from "next/link";

const POSITIONING_TAGS = [
  "传统文化学习",
  "自我观察",
  "行动节奏整理",
] as const;

const HOME_DISCLAIMER =
  "仅供娱乐、传统文化学习和自我反思参考，不构成预测、医疗、法律、投资或重大决策建议。";

const NAV_ENTRIES = [
  {
    href: "/ask",
    title: "问事起卦",
    variant: "primary" as const,
  },
  {
    href: "/bazi",
    title: "八字简析",
    subtitle: "基于传统干支文化的性格与五行倾向参考",
    variant: "bazi" as const,
    deco: "orbit" as const,
  },
  {
    href: "/qimen",
    title: "奇门问事",
    subtitle: "基于传统奇门文化的局势梳理与行动节奏参考",
    variant: "qimen" as const,
    deco: "scan" as const,
  },
  {
    href: "/today",
    title: "今日一卦",
    variant: "today" as const,
  },
  {
    href: "/history",
    title: "历史记录",
    variant: "default" as const,
  },
  {
    href: "/about",
    title: "关于与免责声明",
    variant: "default" as const,
  },
] as const;

export default function Home() {
  return (
    <main className="home-page mx-auto w-full max-w-lg flex-1 px-4 py-6 sm:max-w-xl sm:py-10">
      <header className="home-hero">
        <p className="home-eyebrow">文易传统文化</p>
        <h1 className="home-title">易经问事、八字简析、奇门问事</h1>
        <p className="home-description">
          传统文化学习参考，不构成现实决策依据。
        </p>
      </header>

      <div className="home-positioning" aria-label="产品定位">
        {POSITIONING_TAGS.map((tag) => (
          <div key={tag} className="home-positioning-item">
            <span className="home-positioning-dot" aria-hidden />
            <span>{tag}</span>
          </div>
        ))}
      </div>

      <nav className="home-nav" aria-label="功能入口">
        {NAV_ENTRIES.map((entry) => (
          <Link
            key={entry.href}
            href={entry.href}
            className={`home-nav-link home-nav-link--${entry.variant}`}
          >
            {"subtitle" in entry ? (
              <span className="home-nav-link__content">
                {"deco" in entry && entry.deco === "orbit" ? (
                  <span className="home-entry-deco home-entry-deco-wrap" aria-hidden>
                    <ElementOrbit compact />
                  </span>
                ) : null}
                {"deco" in entry && entry.deco === "scan" ? (
                  <span className="home-entry-deco home-entry-deco-wrap" aria-hidden>
                    <QimenScanGrid compact />
                  </span>
                ) : null}
                <span className="home-nav-link__text">
                  <span>{entry.title}</span>
                  <span className="home-nav-subtitle">{entry.subtitle}</span>
                </span>
              </span>
            ) : (
              <span>{entry.title}</span>
            )}
            <span className="home-nav-arrow" aria-hidden>
              →
            </span>
          </Link>
        ))}
      </nav>

      <p className="home-disclaimer" role="note">
        {HOME_DISCLAIMER}
      </p>
    </main>
  );
}
