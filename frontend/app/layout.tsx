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
  title: "文易传统文化",
  description:
    "提供易经问事、八字简析与奇门问事三类传统文化学习工具，帮助整理当前状态、自我观察与行动节奏。",
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
