import type { ReactNode } from "react";
import { ClerkProvider } from "@clerk/nextjs";

export default function Layout({ children }: { children: ReactNode }) {
  const key = process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;
  const content = key ? (
    <ClerkProvider
      publishableKey={key}
      signInUrl="/sign-in"
      signUpUrl="/sign-up"
      appearance={{
        variables: {
          colorPrimary: "#1D4ED8",
          colorForeground: "#0F172A",
          colorMutedForeground: "#475569",
          fontFamily: "system-ui, sans-serif",
          fontSize: "16px",
          borderRadius: "4px",
        },
        elements: {
          card: { boxShadow: "none", maxWidth: "100%" },
          formButtonPrimary: { minHeight: "44px" },
          formFieldInput: { minHeight: "44px" },
        },
      }}
    >
      {children}
    </ClerkProvider>
  ) : (
    <main>
      <h1>Flux</h1>
      <p role="alert">Sign-in is temporarily unavailable. Try again shortly.</p>
      <a href="/sign-in">Retry sign-in</a>
    </main>
  );
  return (
    <html lang="en">
      <body
        style={{
          margin: 0,
          background: "#F1F5F9",
          color: "#0F172A",
          fontFamily: "system-ui, sans-serif",
          fontSize: "16px",
          lineHeight: 1.5,
        }}
      >
        {content}
      </body>
    </html>
  );
}
