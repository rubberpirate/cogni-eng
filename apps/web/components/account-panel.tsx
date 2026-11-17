"use client";
import { useCallback, useEffect, useState } from "react";
import { apiClient, type components } from "../../../clients/http";
type Profile = components["schemas"]["Profile"];
export function AccountPanel({
  csrf,
  onChanged,
}: {
  csrf: string;
  onChanged: (profile: Profile) => void;
}) {
  const [profile, setProfile] = useState<Profile | null>(null),
    [events, setEvents] = useState<components["schemas"]["AccountEvent"][]>([]),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [busy, setBusy] = useState(false);
  const api = useCallback(() => apiClient(csrf), [csrf]);
  const refresh = useCallback(async () => {
    const [p, e] = await Promise.all([
      api().GET("/api/v1/profile"),
      api().GET("/api/v1/account-activity"),
    ]);
    if (p.error || e.error)
      throw new Error(p.error?.message ?? e.error?.message);
    setProfile(p.data);
    setEvents(e.data);
  }, [api]);
  useEffect(() => {
    void refresh().catch((e) =>
      setError(e instanceof Error ? e.message : "Unable to load account."),
    );
  }, [refresh]);
  async function run(
    fn: () => Promise<{ error?: { message: string } }>,
    message: string,
  ) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const r = await fn();
      if (r.error) throw new Error(r.error.message);
      setNotice(message);
      await refresh();
      return true;
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Connection failed. Your input is retained.",
      );
      return false;
    } finally {
      setBusy(false);
    }
  }
  const get = (f: FormData, k: string) => String(f.get(k) ?? "");
  if (!profile)
    return (
      <section className="editor">
        <p role={error ? "alert" : "status"}>{error || "Loading account…"}</p>
        <button
          onClick={() => void refresh().catch((e) => setError(e.message))}
        >
          Retry
        </button>
      </section>
    );
  return (
    <>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {notice && (
        <p className="notice" role="status">
          {notice}
        </p>
      )}
      <section className="editor">
        <h2>Your profile</h2>
        <p>Verified email: {profile.email}</p>
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const f = new FormData(e.currentTarget);
            const next = {
              name: get(f, "name"),
              phone: get(f, "phone"),
              emergency_contact: get(f, "emergency_contact"),
            };
            void run(
              () =>
                api().PATCH("/api/v1/profile", {
                  params: { header: { "X-CSRF-Token": csrf } },
                  body: next,
                }),
              "Profile saved.",
            ).then((ok) => {
              if (ok) onChanged({ ...profile, ...next });
            });
          }}
        >
          <label>
            Name
            <input
              name="name"
              required
              defaultValue={profile.name}
              maxLength={120}
            />
          </label>
          <label>
            Phone (not verified)
            <input
              name="phone"
              type="tel"
              defaultValue={profile.phone}
              maxLength={40}
            />
          </label>
          <label>
            Emergency contact
            <input
              name="emergency_contact"
              defaultValue={profile.emergency_contact}
              maxLength={300}
            />
          </label>
          <div className="form-actions">
            <button className="primary" disabled={busy}>
              Save profile
            </button>
          </div>
        </form>
      </section>
      <section className="editor">
        <h2>Change password</h2>
        <p>
          Other sessions and unused reset links will be revoked. This session
          stays active.
        </p>
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const form = e.currentTarget,
              f = new FormData(form);
            void run(
              () =>
                api().POST("/api/v1/password", {
                  params: { header: { "X-CSRF-Token": csrf } },
                  body: {
                    current_password: get(f, "current_password"),
                    password: get(f, "password"),
                    code: get(f, "code"),
                  },
                }),
              "Password changed; other sessions ended.",
            ).then((ok) => {
              if (ok) form.reset();
            });
          }}
        >
          <label>
            Current password
            <input
              name="current_password"
              type="password"
              autoComplete="current-password"
              required
              maxLength={256}
            />
          </label>
          <label>
            New password
            <input
              name="password"
              type="password"
              autoComplete="new-password"
              required
              minLength={14}
              maxLength={256}
            />
          </label>
          <label>
            Authenticator or recovery code (if enabled)
            <input name="code" autoComplete="one-time-code" maxLength={32} />
          </label>
          <div className="form-actions">
            <button disabled={busy}>Change password</button>
          </div>
        </form>
      </section>
      <section className="editor">
        <h2>Change verified email</h2>
        <p>
          The current address remains active until you confirm the link sent to
          the new address. In development, use the local mail catcher.
        </p>
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const form = e.currentTarget,
              f = new FormData(form);
            void run(
              () =>
                api().POST("/api/v1/email-change", {
                  params: { header: { "X-CSRF-Token": csrf } },
                  body: {
                    email: get(f, "email"),
                    password: get(f, "password"),
                    code: get(f, "code"),
                  },
                }),
              "Confirmation queued. Your email has not changed yet.",
            ).then((ok) => {
              if (ok) form.reset();
            });
          }}
        >
          <label>
            New email
            <input type="email" name="email" required maxLength={254} />
          </label>
          <label>
            Current password
            <input
              name="password"
              type="password"
              autoComplete="current-password"
              required
              maxLength={256}
            />
          </label>
          <label>
            Authenticator or recovery code (if enabled)
            <input name="code" autoComplete="one-time-code" maxLength={32} />
          </label>
          <div className="form-actions">
            <button disabled={busy}>Send confirmation</button>
          </div>
        </form>
      </section>
      <section className="ledger">
        <h2>Account activity</h2>
        {events.length === 0 ? (
          <p>No account events yet.</p>
        ) : (
          <ol className="activity-list">
            {events.map((e) => (
              <li key={e.id}>
                <strong>{e.action.replaceAll(".", " ")}</strong>
                <time dateTime={e.time}>
                  {new Date(e.time).toLocaleString()}
                </time>
                {e.reason && <p>{e.reason}</p>}
              </li>
            ))}
          </ol>
        )}
      </section>
    </>
  );
}
