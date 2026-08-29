import { SignIn } from '@clerk/clerk-react';

export function SignInPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-950 p-4">
      <SignIn routing="path" path="/login" signUpUrl="/signup" forceRedirectUrl="/dashboard" />
    </div>
  );
}
