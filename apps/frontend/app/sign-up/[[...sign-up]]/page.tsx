"use client";

import { ClerkFailed, ClerkLoaded, ClerkLoading, Show, SignUp, UserButton } from "@clerk/nextjs";

export default function SignUpPage() {
  return (
    <main
      style={{ maxWidth: "448px", margin: "48px auto", padding: "24px", boxSizing: "border-box" }}
    >
      <p>Flux</p>
      <h1 style={{ fontSize: "28px", fontWeight: 600 }}>Create your account</h1>
      <ClerkLoading>
        <p role="status">Loading sign-up…</p>
      </ClerkLoading>
      <ClerkFailed>
        <p role="alert">Sign-in is temporarily unavailable. Try again shortly.</p>
        <a href="/sign-up">Retry sign-in</a>
      </ClerkFailed>
      <ClerkLoaded>
        <Show when="signed-out">
          <SignUp routing="path" path="/sign-up" signInUrl="/sign-in" forceRedirectUrl="/sign-in" />
        </Show>
        <Show when="signed-in">
          <p role="status">Your account is signed in.</p>
          <UserButton />
        </Show>
      </ClerkLoaded>
    </main>
  );
}
