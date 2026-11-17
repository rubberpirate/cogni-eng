import Link from "next/link";
export default function Page() {
  return (
    <main id="main" className="public prose">
      <Link href="/">Learning Workspace</Link>
      <h1>Getting into your workspace</h1>
      <p>
        Your institution invites you by email. Open the single-use invitation,
        set a password of at least 14 characters and acknowledge the development
        policies. Existing accounts keep their own password when another
        institution adds a responsibility.
      </p>
      <h2>Lost access</h2>
      <p>
        Use password recovery from sign in. If authenticator protection is
        enabled, you also need your authenticator or an unused recovery code. If
        both are lost, contact your institution; automatic bypass is
        unavailable.
      </p>
      <h2>Local demonstration</h2>
      <p>
        The startup command writes a random synthetic owner's password to{" "}
        <code>.local/demo-access.txt</code>. Invitations and recovery mail go to
        the local mail catcher on port 58025. No actual email delivery is
        claimed.
      </p>
      <Link href="/sign-in">Sign in</Link>
    </main>
  );
}
