import { SignUp } from '@clerk/clerk-react';

export function SignUpPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-950 p-4">
      <SignUp routing="path" path="/signup" signInUrl="/login" forceRedirectUrl="/dashboard" />
    </div>
  );
}
