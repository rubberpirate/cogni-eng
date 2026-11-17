import Link from "next/link";
export default function Page() {
  return (
    <main id="main" className="public prose">
      <Link href="/">Learning Workspace</Link>
      <h1>Development terms and privacy</h1>
      <p>
        Policy versions: terms-v1 and privacy-v1. This development installation
        is for synthetic data and evaluation. The institution owner must supply
        approved operational terms, retention, consent and jurisdiction details
        before external use.
      </p>
      <h2>Information and access</h2>
      <p>
        The workspace stores account details, institution memberships, placement
        information and security audit events. Your institution controls access
        by campus and responsibility. Passwords and recovery codes are hashed;
        authenticator secrets are encrypted. Sessions can be revoked in Access
        and security.
      </p>
      <h2>Records and requests</h2>
      <p>
        Removing access retains history. Contact your institution about data
        corrections and retention. Formal privacy-request tooling is still
        unimplemented; no automatic legal compliance is claimed.
      </p>
    </main>
  );
}
