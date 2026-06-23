import Link from "next/link";
import { DISCLAIMER } from "./DisclaimerBanner";

export default function Footer() {
  return (
    <footer className="mt-auto border-t border-stone-200 bg-stone-50 px-4 py-6 text-center text-xs leading-relaxed text-stone-500">
      <p>{DISCLAIMER}</p>
      <p className="mt-2">
        <Link href="/about" className="text-amber-800 hover:underline">
          关于与合规说明
        </Link>
        <span className="mx-2 text-stone-300">·</span>
        本地 MVP · mock 解读与解锁
      </p>
    </footer>
  );
}
