import type { Metadata } from "next";
import { Hanken_Grotesk, Noto_Sans_KR } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import "./globals.css";
import { Header } from "@/components/layout/Header";
import { QueryProvider } from "@/components/providers/QueryProvider";

export const metadata: Metadata = {
  title: "PokéTrack Korea",
  description: "Pokémon TCG price tracker across Korean marketplaces",
};

const hankenGrotesk = Hanken_Grotesk({
  subsets: ["latin"],
  variable: "--font-hanken",
});

const notoSansKR = Noto_Sans_KR({
  subsets: ["latin"],
  variable: "--font-noto-kr",
});

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const messages = await getMessages();

  return (
    <html lang="ko" className={`${hankenGrotesk.variable} ${notoSansKR.variable}`}>
      <body className="bg-[#0c0c11]">
        <NextIntlClientProvider messages={messages}>
          <QueryProvider>
            <Header />
            <div className="min-h-screen bg-[#0c0c11]">{children}</div>
            <div
              className="fixed -top-[12%] -left-[6%] w-[48vw] h-[48vw] rounded-full blur-[20px] pointer-events-none
              bg-[radial-gradient(circle,rgba(124,58,237,.12),transparent_62%)]"
            />
            <div
              className="fixed -bottom-[18%] -right-[8%] w-[46vw] h-[46vw] rounded-full blur-[20px] pointer-events-none
              bg-[radial-gradient(circle,rgba(42,111,219,.1),transparent_62%)]"
            />
          </QueryProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
