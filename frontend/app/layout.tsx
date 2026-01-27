// app/layout.tsx
import "antd/dist/reset.css";
import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Oldal statisztikák",
  description: "A bnbdevelopment oldalainak statisztikái",
  icons: {
    icon: "/logo.png",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="hu" suppressHydrationWarning>
      <body suppressHydrationWarning>{children}</body>
    </html>
  );
}
