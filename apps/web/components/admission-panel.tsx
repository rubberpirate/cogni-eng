"use client";
import { FormField } from "./form-field";
import { useCallback, useEffect, useState } from "react";
import { apiClient, type components } from "../../../clients/http";
type Campus = components["schemas"]["Campus"];
type Course = components["schemas"]["Course"];
type Admission = components["schemas"]["Admission"];
type Code = components["schemas"]["AdmissionCode"];
export function AdmissionPanel({
  org,
  campus,
  csrf,
  campuses,
  timezone,
}: {
  org: string;
  campus: string;
  csrf: string;
  campuses: Campus[];
  timezone: string;
}) {
  const [queue, setQueue] = useState<Admission[]>([]),
    [codes, setCodes] = useState<Code[]>([]),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [created, setCreated] = useState(""),
    [busy, setBusy] = useState(false);
  const refresh = useCallback(async () => {
    const api = apiClient(csrf);
    const params = { path: { org }, query: { campus } };
    const [q, c] = await Promise.all([
      api.GET("/api/v1/institutions/{org}/admission-queue", { params }),
      api.GET("/api/v1/institutions/{org}/admission-codes", { params }),
    ]);
    if (q.error || c.error)
      throw new Error(q.error?.message ?? c.error?.message);
    setQueue(q.data);
    setCodes(c.data);
  }, [org, campus, csrf]);
  useEffect(() => {
    setQueue([]);
    setCodes([]);
    setError("");
    void refresh().catch((e) => setError(e.message));
  }, [refresh]);
  return (
    <>
      {error && (
        <p className="error" role="alert">
          {error}{" "}
          <button
            onClick={() => void refresh().catch((e) => setError(e.message))}
          >
            Retry
          </button>
        </p>
      )}
      {notice && (
        <p role="status" className="notice">
          {notice}
        </p>
      )}
      <section className="editor">
        <h2>Enrollment codes</h2>
        <p>
          Share a code with prospective learners. Verified registrations enter
          the placement queue; codes never grant staff permissions.
        </p>
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const f = new FormData(e.currentTarget);
            setBusy(true);
            setError("");
            setCreated("");
            void apiClient(csrf)
              .POST("/api/v1/institutions/{org}/admission-codes", {
                params: { path: { org }, header: { "X-CSRF-Token": csrf } },
                body: {
                  campus_id: String(f.get("campus")),
                  label: String(f.get("label")),
                  max_uses: Number(f.get("max")),
                  expires_at: new Date(
                    Date.now() + Number(f.get("days")) * 86400000,
                  ).toISOString(),
                },
              })
              .then(async (r) => {
                if (r.error) throw new Error(r.error.message);
                setCreated(r.data.code);
                setNotice(
                  "Enrollment code created. Copy it now; only its hash is retained.",
                );
                await refresh();
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          <FormField name="Campus">
            <select name="campus" required defaultValue={campus}>
              <option value="">Choose campus</option>
              {campuses.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </FormField>
          <label>
            Label
            <input name="label" required minLength={2} maxLength={100} />
          </label>
          <label>
            Valid for days
            <input
              name="days"
              type="number"
              min={1}
              max={30}
              defaultValue={7}
              required
            />
          </label>
          <label>
            Maximum registrations
            <input
              name="max"
              type="number"
              min={1}
              max={1000}
              defaultValue={30}
              required
            />
          </label>
          <div className="form-actions">
            <button className="primary" disabled={busy}>
              Create enrollment code
            </button>
          </div>
        </form>
        {created && (
          <div className="notice">
            <p>Registration code (shown once)</p>
            <code className="secret">{created}</code>
            <p>
              Enrollment page: <a href="/register">/register</a>
            </p>
            <button onClick={() => setCreated("")}>I copied the code</button>
          </div>
        )}
        <ul className="activity-list">
          {codes.map((c) => (
            <li key={c.id}>
              <strong>{c.label}</strong>
              <p>
                {c.used} / {c.max_uses} registrations ·{" "}
                {c.revoked
                  ? "Revoked"
                  : new Intl.DateTimeFormat("en-IN", {
                      dateStyle: "medium",
                      timeStyle: "short",
                      timeZone: timezone,
                    }).format(new Date(c.expires_at))}
              </p>
              {!c.revoked && (
                <button
                  disabled={busy}
                  onClick={() => {
                    if (!confirm("Revoke this code for new registrations?"))
                      return;
                    setBusy(true);
                    void apiClient(csrf)
                      .DELETE(
                        "/api/v1/institutions/{org}/admission-codes/{id}",
                        {
                          params: {
                            path: { org, id: c.id },
                            header: { "X-CSRF-Token": csrf },
                          },
                        },
                      )
                      .then(async (r) => {
                        if (r.error) throw new Error(r.error.message);
                        await refresh();
                      })
                      .catch((e) => setError(e.message))
                      .finally(() => setBusy(false));
                  }}
                >
                  Revoke code
                </button>
              )}
            </li>
          ))}
        </ul>
      </section>
      <section className="editor">
        <h2>Placement queue</h2>
        <p>
          Only a verified pending request can be placed. Rejected and already
          placed requests retain their history.
        </p>
        {queue.length === 0 ? (
          <p>No admission requests in this context.</p>
        ) : (
          queue.map((q) => (
            <AdmissionDecision
              key={`${q.id}-${q.version}`}
              admission={q}
              org={org}
              csrf={csrf}
              onSaved={() => void refresh().catch((e) => setError(e.message))}
            />
          ))
        )}
      </section>
    </>
  );
}
function AdmissionDecision({
  admission,
  org,
  csrf,
  onSaved,
}: {
  admission: Admission;
  org: string;
  csrf: string;
  onSaved: () => void;
}) {
  const [courses, setCourses] = useState<Course[]>([]),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  useEffect(() => {
    if (admission.state !== "pending") return;
    void apiClient(csrf)
      .GET("/api/v1/institutions/{org}/courses", {
        params: { path: { org }, query: { campus: admission.campus_id } },
      })
      .then((r) => {
        if (r.error) setError(r.error.message);
        else setCourses(r.data);
      })
      .catch(() => setError("Could not load courses."));
  }, [admission.campus_id, admission.state, org, csrf]);
  return (
    <article className="queue-item">
      <h3>{admission.name}</h3>
      <p>
        {admission.email} · {admission.state}
      </p>
      {admission.reason && <p>{admission.reason}</p>}
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {admission.state === "pending" && (
        <form
          className="field-grid"
          onSubmit={(e) => {
            e.preventDefault();
            const f = new FormData(e.currentTarget);
            if (!confirm("Apply this admission decision?")) return;
            setBusy(true);
            void apiClient(csrf)
              .POST(
                "/api/v1/institutions/{org}/admission-queue/{id}/decision",
                {
                  params: {
                    path: { org, id: admission.id },
                    header: { "X-CSRF-Token": csrf },
                  },
                  body: {
                    course_id: String(f.get("course")),
                    start_date: String(f.get("date")),
                    decision: String(f.get("decision")) as
                      "placed" | "rejected",
                    reason: String(f.get("reason")),
                    version: admission.version,
                  },
                },
              )
              .then((r) => {
                if (r.error) throw new Error(r.error.message);
                onSaved();
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          <FormField name="Decision">
            <select name="decision">
              <option value="placed">Place in course</option>
              <option value="rejected">Reject request</option>
            </select>
          </FormField>
          <FormField name="Course (for placement)">
            <select name="course">
              <option value="">Choose course</option>
              {courses.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </FormField>
          <label>
            Start date (for placement)
            <input name="date" type="date" />
          </label>
          <label>
            Decision reason
            <input name="reason" required minLength={5} maxLength={500} />
          </label>
          <div className="form-actions">
            <button disabled={busy}>Apply decision</button>
          </div>
        </form>
      )}
    </article>
  );
}
