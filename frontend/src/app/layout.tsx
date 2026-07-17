import type { Metadata } from "next";

import "./globals.css";

export const metadata: Metadata = {
  title: "Kyver",
  description: "AI-powered software engineering platform foundation",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased">{children}</body>
    </html>
  );
}
