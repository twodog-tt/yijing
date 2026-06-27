import Link from "next/link";

const HOME_BRAND = {
  eyebrow: "文易传统文化",
  title: "易经问事、八字简析、奇门问事",
  subtitle: "传统文化学习参考，不构成现实决策依据",
} as const;

const HOME_COMPLIANCE_NOTE =
  "本产品内容基于传统文化模型生成，仅供学习参考、自我观察与行动节奏整理。结果不构成现实决策依据，请结合实际情况判断。";

const HOME_MODULES = [
  {
    href: "/ask",
    title: "问事起卦",
    subtitle: "适合梳理一个具体问题的当前状态与行动提醒",
    tags: ["具体问题", "局势观察", "行动提醒"],
    buttonText: "开始问事",
    accent: "ask",
  },
  {
    href: "/bazi",
    title: "八字简析",
    subtitle: "适合从传统文化视角观察个人结构与长期节奏",
    tags: ["自我观察", "五行结构", "长期节奏"],
    buttonText: "查看八字",
    accent: "bazi",
  },
  {
    href: "/qimen",
    title: "奇门问事",
    subtitle: "适合观察当前局势、资源关系与推进节奏",
    tags: ["局势梳理", "资源关系", "推进节奏"],
    buttonText: "进入奇门",
    accent: "qimen",
  },
] as const;

const HOME_GUIDE_ITEMS = [
  { scenario: "临时问题 / 具体事情", module: "问事起卦" },
  { scenario: "自我结构 / 长期节奏", module: "八字简析" },
  { scenario: "当前局势 / 行动节奏", module: "奇门问事" },
] as const;

const HOME_SCENE_ITEMS = [
  {
    href: "/ask",
    title: "感情关系观察",
    subtitle: "适合梳理关系状态、沟通节奏与边界提醒",
    description: "遇到感情困惑时，可以先从一个具体问题开始观察。",
    buttonText: "去问感情问题",
  },
] as const;

const HOME_BOUNDARY_ITEMS = [
  "不做精准预测",
  "不替代现实决策",
  "不提供投资、医疗、法律等建议",
] as const;

export default function Home() {
  return (
    <main className="home-page flex-1">
      <section className="home-hero">
        <div className="home-hero__aura" aria-hidden />
        <p className="home-hero__eyebrow">{HOME_BRAND.eyebrow}</p>
        <h1 className="home-hero__title">{HOME_BRAND.title}</h1>
        <p className="home-hero__subtitle">{HOME_BRAND.subtitle}</p>
      </section>

      <section className="home-modules" aria-label="核心模块">
        {HOME_MODULES.map((module) => (
          <Link
            key={module.href}
            href={module.href}
            className={`home-module-card home-module-card--${module.accent}`}
          >
            <span className="home-module-card__accent" aria-hidden />
            <span className="home-module-card__header">
              <span className="home-module-card__title">{module.title}</span>
              <span className="home-module-card__subtitle">{module.subtitle}</span>
            </span>
            <span className="home-module-card__tags">
              {module.tags.map((tag) => (
                <span key={tag} className="home-module-card__tag">
                  {tag}
                </span>
              ))}
            </span>
            <span className="home-module-card__action">
              <span className="home-module-card__button">{module.buttonText}</span>
              <span className="home-module-card__arrow" aria-hidden>
                →
              </span>
            </span>
          </Link>
        ))}
      </section>

      <section className="home-guide-card">
        <h2 className="home-guide-card__title">如何选择模块</h2>
        <div className="home-guide-card__list">
          {HOME_GUIDE_ITEMS.map((item) => (
            <div key={item.scenario} className="home-guide-card__item">
              <span className="home-guide-card__scenario">{item.scenario}</span>
              <span className="home-guide-card__module">{item.module}</span>
            </div>
          ))}
        </div>
      </section>

      <section className="home-scenes">
        <h2 className="home-scenes__title">常见场景</h2>
        {HOME_SCENE_ITEMS.map((item) => (
          <Link key={item.title} href={item.href} className="home-scene-card">
            <span className="home-scene-card__main">
              <span className="home-scene-card__title">{item.title}</span>
              <span className="home-scene-card__subtitle">{item.subtitle}</span>
              <span className="home-scene-card__description">
                {item.description}
              </span>
            </span>
            <span className="home-scene-card__button">{item.buttonText}</span>
          </Link>
        ))}
      </section>

      <section className="home-boundary">
        <h2 className="home-boundary__title">使用边界说明</h2>
        <div className="home-boundary__list">
          {HOME_BOUNDARY_ITEMS.map((item) => (
            <p key={item} className="home-boundary__item">
              · {item}
            </p>
          ))}
        </div>
      </section>

      <nav className="home-secondary" aria-label="次级入口">
        <Link href="/history" className="home-secondary__link">
          <span>历史记录</span>
          <span className="home-secondary__arrow" aria-hidden>
            →
          </span>
        </Link>
        <Link href="/about" className="home-secondary__link">
          <span>关于与说明</span>
          <span className="home-secondary__arrow" aria-hidden>
            →
          </span>
        </Link>
      </nav>

      <p className="home-compliance" role="note">
        {HOME_COMPLIANCE_NOTE}
      </p>
    </main>
  );
}
