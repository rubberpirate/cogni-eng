"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { apiClient } from "../../../clients/http";
export function AccessForm({ mode }: { mode: "sign-in" | "reset" | "join" }) {
  const [error, setError] = useState(""),
    [message, setMessage] = useState(""),
    [busy, setBusy] = useState(false),
    [show, setShow] = useState(false),
    [recovery, setRecovery] = useState(false),
    [token, setToken] = useState("");
  useEffect(() => {
    const incoming = location.hash.slice(1);
    if (incoming) {
      setToken(incoming);
      history.replaceState(null, "", location.pathname);
    }
  }, []);
  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setMessage("");
    setBusy(true);
    const f = new FormData(e.currentTarget);
    const api = apiClient();
    try {
      if (recovery) {
        const result = await api.POST("/api/v1/recovery", {
          body: { email: String(f.get("email")) },
        });
        if (result.error) throw new Error(result.error.message);
        setMessage(
          "If the account is eligible, recovery instructions will be sent. Development mail appears in the local mail catcher.",
        );
      } else if (mode === "sign-in") {
        const result = await api.POST("/api/v1/sign-in", {
          body: {
            email: String(f.get("email")),
            password: String(f.get("password")),
            code: String(f.get("code") || ""),
            remember: f.get("remember") === "on",
          },
        });
        if (result.error) throw new Error(result.error.message);
        location.assign("/workspace");
      } else {
        const body = {
          token,
          password: String(f.get("password")),
          code: String(f.get("code") || ""),
          accept_policy: f.get("policy") === "on",
        };
        const result =
          mode === "join"
            ? await api.POST("/api/v1/accept-invitation", { body })
            : await api.POST("/api/v1/reset", { body });
        if (result.error) throw new Error(result.error.message);
        setMessage("Your password is set. Sign in to open your workspace.");
      }
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Connection failed. Your input is retained; try again.",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <main id="main" className="access">
      <Link href="/" className="wordmark">
        LW / Learning Workspace
      </Link>
      <section className="access-panel">
        <p className="eyebrow">Your institution. Your responsibilities.</p>
        <h1>
          {recovery
            ? "Recover access"
            : mode === "sign-in"
              ? "Welcome back"
              : mode === "join"
                ? "Accept your invitation"
                : "Set a new password"}
        </h1>
        <p>
          {mode === "sign-in"
            ? "Use the email associated with your institution."
            : "Links expire after use. Choose a unique password of at least 14 characters."}
        </p>
        <form onSubmit={submit} aria-busy={busy}>
          {(mode === "sign-in" || recovery) && (
            <label>
              Email address
              <input
                type="email"
                name="email"
                autoComplete="username"
                required
                maxLength={254}
              />
            </label>
          )}
          {!recovery && (
            <>
              <div>
                <label htmlFor="password">Password</label>
                <div className="input-action">
                  <input
                    id="password"
                    type={show ? "text" : "password"}
                    name="password"
                    autoComplete={
                      mode === "sign-in" ? "current-password" : "new-password"
                    }
                    required
                    minLength={mode === "sign-in" ? 1 : 14}
                    maxLength={256}
                  />
                  <button
                    type="button"
                    aria-pressed={show}
                    onClick={() => setShow(!show)}
                  >
                    {show ? "Hide" : "Show"}
                  </button>
                </div>
              </div>
              <label>
                Authenticator or recovery code{" "}
                <span className="muted">(if enabled)</span>
                <input
                  name="code"
                  autoComplete="one-time-code"
                  maxLength={32}
                />
              </label>
            </>
          )}
          {mode === "sign-in" && !recovery && (
            <label className="check">
              <input type="checkbox" name="remember" />
              Keep this device signed in (30 days, subject to idle expiry)
            </label>
          )}
          {mode === "join" && (
            <label className="check">
              <input type="checkbox" name="policy" required />I accept the{" "}
              <Link href="/policies">development terms and privacy policy</Link>
              .
            </label>
          )}
          {error && (
            <p role="alert" className="error">
              {error}
            </p>
          )}
          {message && (
            <p role="status" className="notice">
              {message}
            </p>
          )}
          <button
            className="primary"
            disabled={busy || (!token && mode !== "sign-in")}
          >
            {busy
              ? "Working…"
              : recovery
                ? "Request recovery"
                : mode === "sign-in"
                  ? "Sign in"
                  : "Set password"}
          </button>
        </form>
        {mode === "sign-in" ? (
          <button
            className="text-button"
            onClick={() => {
              setRecovery(!recovery);
              setError("");
              setMessage("");
            }}
          >
            {recovery ? "Return to sign in" : "Forgot your password?"}
          </button>
        ) : (
          <Link href="/sign-in">Return to sign in</Link>
        )}
        <p className="muted">
          New learner? <Link href="/register">Request enrollment</Link>
          <br />
          Need a hand? <Link href="/help">Access help</Link>
        </p>
      </section>
    </main>
  );
}
