import Link from "next/link";

export default function Footer() {
  return (
    <footer className="mt-auto border-t border-stone-200/80 bg-stone-50/80 px-4 py-3 text-center text-xs leading-relaxed text-stone-500 sm:py-4">
      <p>
        <Link href="/about" className="inline-flex min-h-9 items-center text-amber-800 hover:underline">
          关于与免责声明
        </Link>
        <span className="mx-2 text-stone-300">·</span>
        <span className="text-stone-400">AI 解读 · 内测体验</span>
      </p>
    </footer>
  );
}
