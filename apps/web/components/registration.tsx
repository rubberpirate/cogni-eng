"use client";
import Link from "next/link";
import { useState } from "react";
import { apiClient } from "../../../clients/http";
export function Registration() {
  const [error, setError] = useState(""),
    [message, setMessage] = useState(""),
    [busy, setBusy] = useState(false);
  return (
    <main id="main" className="access">
      <Link className="wordmark" href="/">
        LW / Learning Workspace
      </Link>
      <section className="access-panel">
        <h1>Join as a learner</h1>
        <p>
          Your institution supplies an enrollment code. Verify your email, then
          staff will review your placement.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            const form = e.currentTarget,
              f = new FormData(form);
            setBusy(true);
            setError("");
            void apiClient()
              .POST("/api/v1/register", {
                body: {
                  name: String(f.get("name")),
                  email: String(f.get("email")),
                  password: String(f.get("password")),
                  enrollment_code: String(f.get("enrollment_code")),
                  accept_policy: f.get("policy") === "on",
                },
              })
              .then((r) => {
                if (r.error) throw new Error(r.error.message);
                setMessage(
                  r.data.message ??
                    "Check your email for verification instructions.",
                );
                form.reset();
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          <label>
            Name
            <input name="name" required maxLength={120} autoComplete="name" />
          </label>
          <label>
            Email address
            <input
              name="email"
              type="email"
              required
              maxLength={254}
              autoComplete="email"
            />
          </label>
          <label>
            New password
            <input
              name="password"
              type="password"
              required
              minLength={14}
              maxLength={256}
              autoComplete="new-password"
            />
          </label>
          <label>
            Enrollment code
            <input name="enrollment_code" required maxLength={64} />
          </label>
          <label className="check">
            <input name="policy" type="checkbox" required />I accept the{" "}
            <Link href="/policies">development terms and privacy policy</Link>.
          </label>
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
          <button className="primary" disabled={busy}>
            Request enrollment
          </button>
        </form>
        <p>
          Already have an account? <Link href="/sign-in">Sign in</Link> and
          contact your institution for an additional membership.
        </p>
        <Link href="/verify">Resend email verification</Link>
      </section>
    </main>
  );
}
