import Link from "next/link";
import { DISCLAIMER } from "./DisclaimerBanner";

export default function Footer() {
  return (
    <footer className="mt-auto border-t border-stone-200 bg-stone-50 px-4 py-4 text-center text-xs leading-relaxed text-stone-500 sm:py-6">
      <p>{DISCLAIMER}</p>
      <p className="mt-2">
        <Link href="/about" className="inline-flex min-h-11 items-center text-amber-800 hover:underline">
          关于与合规说明
        </Link>
        <span className="mx-2 text-stone-300">·</span>
        AI 解读 · 模拟解锁
      </p>
    </footer>
  );
}
