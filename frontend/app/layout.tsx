import type { Metadata } from "next";
import { Noto_Serif_SC } from "next/font/google";
import DisclaimerBanner from "@/components/layout/DisclaimerBanner";
import Footer from "@/components/layout/Footer";
import "./globals.css";

const notoSerif = Noto_Serif_SC({
  subsets: ["latin"],
  weight: ["400", "600", "700"],
  variable: "--font-serif-sc",
});

export const metadata: Metadata = {
  title: "易经问事 · AI 卦象解读",
  description:
    "AI 易经问事与卦象解读工具，仅供娱乐和传统文化参考，助力自我反思与行动建议。",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" className={`${notoSerif.variable} h-full`}>
      <body className="flex min-h-dvh flex-col bg-[#faf8f3] font-serif text-stone-900 antialiased">
        <DisclaimerBanner />
        <div className="flex flex-1 flex-col">{children}</div>
        <Footer />
      </body>
    </html>
  );
}
