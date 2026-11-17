"use client";
export default function WorkspaceError({ reset }: { reset: () => void }) {
  return (
    <main id="main" className="public prose">
      <h1>Your workspace could not load</h1>
      <p>
        The service could not provide a permitted workspace. Your saved work
        remains on the server. Retry when the connection recovers.
      </p>
      <button onClick={reset}>Retry workspace</button>
      <p>
        <a href="/sign-in">Return to sign in</a>
      </p>
    </main>
  );
}
