import NextAuth from "next-auth";
import { authConfig } from "@/lib/auth.config";
import { NextResponse } from "next/server";

const { auth } = NextAuth(authConfig);

export default auth((req) => {
  const { nextUrl } = req;
  const isLoggedIn = !!req.auth?.user;
  const role = (req.auth?.user as { role?: string } | undefined)?.role;
  const path = nextUrl.pathname;

  const isLogin = path === "/login";
  const isAdminArea = path.startsWith("/admin");

  if (!isLoggedIn && !isLogin) {
    return NextResponse.redirect(new URL("/login", nextUrl));
  }
  if (isLoggedIn && isLogin) {
    return NextResponse.redirect(new URL(role === "ADMIN" ? "/admin" : "/jobs", nextUrl));
  }
  if (isAdminArea && role !== "ADMIN") {
    return NextResponse.redirect(new URL("/jobs", nextUrl));
  }
  return NextResponse.next();
});

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
