import { signIn } from "@/lib/auth";
import { AuthError } from "next-auth";
import { redirect } from "next/navigation";

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const params = await searchParams;
  return (
    <div style={{ maxWidth: 360, margin: "10vh auto" }}>
      <h1>Klaus Field Log</h1>
      {params.error && <p style={{ color: "red" }}>Invalid email or password.</p>}
      <form
        action={async (formData: FormData) => {
          "use server";
          try {
            await signIn("credentials", {
              email: formData.get("email"),
              password: formData.get("password"),
              redirectTo: "/jobs",
            });
          } catch (error) {
            if (error instanceof AuthError) {
              redirect("/login?error=CredentialsSignin");
            }
            throw error;
          }
        }}
        style={{ display: "flex", flexDirection: "column", gap: 10 }}
      >
        <input name="email" type="email" placeholder="Email" required />
        <input name="password" type="password" placeholder="Password" required />
        <button type="submit">Log in</button>
      </form>
    </div>
  );
}
