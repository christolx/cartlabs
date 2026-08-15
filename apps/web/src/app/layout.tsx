import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { SessionProvider } from "@/components/session-provider";
import { SiteFooter } from "@/components/site-footer";
import { SiteHeader } from "@/components/site-header";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: { default: "Cartlabs Market", template: "%s | Cartlabs" },
  description: "Verified independent stores, live SKU inventory, and a tested multi-vendor marketplace.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  const demoEnabled = process.env.DEMO_MODE === "true" || process.env.NEXT_PUBLIC_DEMO_MODE === "true";
  return (
    <html
      lang="id"
      data-scroll-behavior="smooth"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <SessionProvider demoEnabled={demoEnabled}>
          <SiteHeader />
          <div className="page-frame">{children}</div>
          <SiteFooter />
        </SessionProvider>
      </body>
    </html>
  );
}
