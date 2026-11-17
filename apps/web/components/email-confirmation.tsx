"use client";
import Link from "next/link";
import { useEffect, useState } from "react";
import { apiClient } from "../../../clients/http";
export function EmailConfirmation() {
  const [token, setToken] = useState(""),
    [error, setError] = useState(""),
    [done, setDone] = useState(false),
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
      <Link href="/workspace">Learning Workspace</Link>
      <h1>Confirm your new email</h1>
      <p>
        Sign in to the account that requested this change. The link is
        single-use and expires after one hour.
      </p>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {done ? (
        <p role="status">
          Your email is now verified. Other sessions have ended.
        </p>
      ) : (
        <button
          className="primary"
          disabled={!token || busy}
          onClick={() => {
            setBusy(true);
            setError("");
            void (async () => {
              try {
                const me = await apiClient().GET("/api/v1/me");
                if (me.error)
                  throw new Error("Sign in first, then reopen the email link.");
                const r = await apiClient(me.data.csrf).POST(
                  "/api/v1/email-change/confirm",
                  {
                    params: { header: { "X-CSRF-Token": me.data.csrf } },
                    body: { token },
                  },
                );
                if (r.error) throw new Error(r.error.message);
                setDone(true);
                setToken("");
              } catch (e) {
                setError(
                  e instanceof Error
                    ? e.message
                    : "Connection failed. Retry with the same link.",
                );
              } finally {
                setBusy(false);
              }
            })();
          }}
        >
          Confirm email
        </button>
      )}
      <p>
        <Link href="/sign-in">Sign in</Link> ·{" "}
        <Link href="/workspace">Return to workspace</Link>
      </p>
    </main>
  );
}
