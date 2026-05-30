import type { ReactNode } from "react";
import Nav from "@/components/Nav";

export const metadata = { title: "Klaus Field Log" };

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body style={{ fontFamily: "system-ui, sans-serif", margin: 0 }}>
        <Nav />
        <main style={{ maxWidth: 720, margin: "0 auto", padding: 16 }}>{children}</main>
      </body>
    </html>
  );
}
