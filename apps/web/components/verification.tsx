"use client";
import Link from "next/link";
import { useEffect, useState } from "react";
import { apiClient } from "../../../clients/http";
export function Verification() {
  const [token, setToken] = useState(""),
    [error, setError] = useState(""),
    [message, setMessage] = useState(""),
    [busy, setBusy] = useState(false);
  useEffect(() => {
    const incoming = location.hash.slice(1);
    if (incoming) {
      setToken(incoming);
      history.replaceState(null, "", location.pathname);
    }
  }, []);
  return (
    <main id="main" className="public prose">
      <Link href="/">Learning Workspace</Link>
      <h1>Verify your email</h1>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {message && (
        <p className="notice" role="status">
          {message}
        </p>
      )}
      {token ? (
        <button
          className="primary"
          disabled={busy}
          onClick={() => {
            setBusy(true);
            setError("");
            void apiClient()
              .POST("/api/v1/verify-email", { body: { token } })
              .then((r) => {
                if (r.error) throw new Error(r.error.message);
                setMessage(
                  "Email verified. Sign in; your institution will place you in a course.",
                );
                setToken("");
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          Verify email
        </button>
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            const f = new FormData(e.currentTarget);
            setBusy(true);
            setError("");
            void apiClient()
              .POST("/api/v1/verification/resend", {
                body: { email: String(f.get("email")) },
              })
              .then((r) => {
                if (r.error) throw new Error(r.error.message);
                setMessage(
                  "If the account is awaiting verification, instructions are queued.",
                );
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          <label>
            Email address
            <input name="email" type="email" required maxLength={254} />
          </label>
          <button disabled={busy}>Resend verification</button>
        </form>
      )}
      <p>
        <Link href="/sign-in">Sign in</Link>
      </p>
    </main>
  );
}
