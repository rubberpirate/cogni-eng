"use client";
import { FormField } from "./form-field";
import { useCallback, useEffect, useState } from "react";
import { apiClient, type components } from "../../../clients/http";
type Settings = components["schemas"]["Settings"];
export function SettingsPanel({
  org,
  csrf,
  onSaved,
}: {
  org: string;
  csrf: string;
  onSaved: () => void;
}) {
  const [settings, setSettings] = useState<Settings | null>(null),
    [staff, setStaff] = useState<components["schemas"]["Person"][]>([]),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [busy, setBusy] = useState(false);
  const refresh = useCallback(async () => {
    const api = apiClient(csrf);
    const [s, p] = await Promise.all([
      api.GET("/api/v1/institutions/{org}/settings", {
        params: { path: { org } },
      }),
      api.GET("/api/v1/institutions/{org}/people", {
        params: { path: { org } },
      }),
    ]);
    if (s.error || p.error)
      throw new Error(s.error?.message ?? p.error?.message);
    setSettings(s.data);
    setStaff(p.data);
  }, [org, csrf]);
  useEffect(() => {
    setSettings(null);
    void refresh().catch((e) => setError(e.message));
  }, [refresh]);
  if (!settings)
    return (
      <section className="editor">
        <p role={error ? "alert" : "status"}>
          {error || "Loading institution settings…"}
        </p>
        <button
          onClick={() => void refresh().catch((e) => setError(e.message))}
        >
          Retry
        </button>
      </section>
    );
  const get = (f: FormData, k: string) => String(f.get(k) ?? "");
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
        <h2>Institution configuration</h2>
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const f = new FormData(e.currentTarget);
            setBusy(true);
            setError("");
            void apiClient(csrf)
              .PATCH("/api/v1/institutions/{org}/settings", {
                params: { path: { org }, header: { "X-CSRF-Token": csrf } },
                body: {
                  ...settings,
                  name: get(f, "name"),
                  timezone: get(f, "timezone"),
                  locale: get(f, "locale") as "en" | "hi",
                  support_email: get(f, "support_email"),
                  academic_year: get(f, "academic_year"),
                  require_staff_mfa: f.get("mfa") === "on",
                },
              })
              .then(async (r) => {
                if (r.error) throw new Error(r.error.message);
                setNotice(
                  "Configuration saved. Policies now apply to staff requests.",
                );
                await refresh();
                onSaved();
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          <label>
            Name
            <input
              name="name"
              required
              defaultValue={settings.name}
              minLength={2}
              maxLength={120}
            />
          </label>
          <label>
            Timezone
            <input name="timezone" required defaultValue={settings.timezone} />
          </label>
          <FormField name="Language">
            <select name="locale" defaultValue={settings.locale}>
              <option value="en">English</option>
              <option value="hi">हिन्दी</option>
            </select>
          </FormField>
          <label>
            Support email
            <input
              name="support_email"
              type="email"
              defaultValue={settings.support_email}
              maxLength={254}
            />
          </label>
          <label>
            Academic year
            <input
              name="academic_year"
              defaultValue={settings.academic_year}
              maxLength={40}
            />
          </label>
          <label className="check">
            <input
              type="checkbox"
              name="mfa"
              defaultChecked={settings.require_staff_mfa}
            />
            Require authenticator protection for staff
          </label>
          <p className="field-hint">
            Staff can sign in and set up their authenticator; institutional
            management requests are denied until setup is complete. Enable your
            own authenticator first.
          </p>
          <div className="form-actions">
            <button className="primary" disabled={busy}>
              Save configuration
            </button>
          </div>
        </form>
      </section>
      <section className="editor">
        <h2>Transfer ownership</h2>
        <p>
          This grants institution-wide ownership to an active, verified staff
          member and ends your owner membership. Other responsibilities remain.
          The action is audited and requires your password and MFA if enabled.
        </p>
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const f = new FormData(e.currentTarget);
            if (
              !confirm(
                "Transfer ownership and relinquish your owner membership?",
              )
            )
              return;
            setBusy(true);
            setError("");
            void apiClient(csrf)
              .POST("/api/v1/institutions/{org}/owner-transfer", {
                params: { path: { org }, header: { "X-CSRF-Token": csrf } },
                body: {
                  person_id: get(f, "person_id"),
                  password: get(f, "password"),
                  code: get(f, "code"),
                  reason: get(f, "reason"),
                },
              })
              .then((r) => {
                if (r.error) throw new Error(r.error.message);
                location.reload();
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          <FormField name="Successor">
            <select name="person_id" required>
              <option value="">Choose verified staff</option>
              {staff
                .filter(
                  (p) =>
                    ["coordinator", "instructor"].includes(p.role) &&
                    p.state === "active" &&
                    p.access_state === "active",
                )
                .map((p) => (
                  <option key={p.membership_id} value={p.id}>
                    {p.name} · {p.email}
                  </option>
                ))}
            </select>
          </FormField>
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
          <label>
            Reason
            <input name="reason" required minLength={10} maxLength={500} />
          </label>
          <div className="form-actions">
            <button disabled={busy}>Transfer ownership</button>
          </div>
        </form>
      </section>
    </>
  );
}
