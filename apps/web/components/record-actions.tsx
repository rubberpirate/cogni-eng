"use client";
import { FormField } from "./form-field";
import { useEffect, useState } from "react";
import { apiClient, type components } from "../../../clients/http";
type Campus = components["schemas"]["Campus"];
type Course = components["schemas"]["Course"];
export function CampusAction({
  org,
  csrf,
  row,
  onSaved,
}: {
  org: string;
  csrf: string;
  row: Campus;
  onSaved: () => void;
}) {
  const [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  return (
    <details>
      <summary>Edit campus</summary>
      <form
        className="inline-editor"
        onSubmit={(e) => {
          e.preventDefault();
          const f = new FormData(e.currentTarget);
          if (
            f.get("state") === "closed" &&
            !confirm("Close this campus for future activity?")
          )
            return;
          setBusy(true);
          setError("");
          void apiClient(csrf)
            .PATCH("/api/v1/institutions/{org}/campuses/{id}", {
              params: {
                path: { org, id: row.id },
                header: { "X-CSRF-Token": csrf },
              },
              body: {
                ...row,
                name: String(f.get("name")),
                code: String(f.get("code")),
                address: String(f.get("address")),
                state: String(f.get("state")),
              },
            })
            .then((r) => {
              if (r.error) throw new Error(r.error.message);
              onSaved();
            })
            .catch((e) => setError(e.message))
            .finally(() => setBusy(false));
        }}
      >
        <label>
          Name
          <input name="name" required defaultValue={row.name} maxLength={120} />
        </label>
        <label>
          Code
          <input name="code" required defaultValue={row.code} maxLength={20} />
        </label>
        <label>
          Address
          <input name="address" defaultValue={row.address} maxLength={500} />
        </label>
        <FormField name="State">
          <select name="state" defaultValue={row.state}>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
          </select>
        </FormField>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <button disabled={busy}>Save campus</button>
      </form>
    </details>
  );
}
export function PlacementAction({
  org,
  csrf,
  row,
  campuses,
  onSaved,
}: {
  org: string;
  csrf: string;
  row: components["schemas"]["Enrollment"];
  campuses: Campus[];
  onSaved: () => void;
}) {
  const [campus, setCampus] = useState(row.campus_id),
    [courses, setCourses] = useState<Course[]>([]),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  useEffect(() => {
    let active = true;
    void apiClient(csrf)
      .GET("/api/v1/institutions/{org}/courses", {
        params: { path: { org }, query: { campus } },
      })
      .then((r) => {
        if (!active) return;
        if (r.error) setError(r.error.message);
        else setCourses(r.data);
      })
      .catch(() => setError("Unable to load courses."));
    return () => {
      active = false;
    };
  }, [campus, csrf, org]);
  return (
    <details>
      <summary>Transfer / withdraw</summary>
      <p>
        The permanent reference is retained. Existing sessions end when
        placement changes.
      </p>
      <form
        className="inline-editor"
        onSubmit={(e) => {
          e.preventDefault();
          const f = new FormData(e.currentTarget);
          if (!confirm("Apply this placement change and retain its history?"))
            return;
          setBusy(true);
          setError("");
          void apiClient(csrf)
            .PATCH("/api/v1/institutions/{org}/enrollments/{id}", {
              params: {
                path: { org, id: row.id },
                header: { "X-CSRF-Token": csrf },
              },
              body: {
                campus_id: campus,
                course_id: String(f.get("course_id")),
                state: String(f.get("state")) as "active" | "withdrawn",
                effective_on: String(f.get("effective_on")),
                reason: String(f.get("reason")),
                version: row.version,
              },
            })
            .then((r) => {
              if (r.error) throw new Error(r.error.message);
              onSaved();
            })
            .catch((e) => setError(e.message))
            .finally(() => setBusy(false));
        }}
      >
        <FormField name="Target campus">
          <select
            value={campus}
            onChange={(e) => setCampus(e.target.value)}
            required
          >
            {campuses.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </FormField>
        <FormField name="Target course">
          <select
            name="course_id"
            required
            key={campus}
            defaultValue={row.course_id}
          >
            <option value="">Choose course</option>
            {courses.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </FormField>
        <FormField name="State">
          <select name="state" defaultValue={row.state}>
            <option value="active">Active</option>
            <option value="withdrawn">Withdrawn</option>
          </select>
        </FormField>
        <label>
          Effective date
          <input name="effective_on" type="date" required />
        </label>
        <label>
          Reason
          <input name="reason" required minLength={5} maxLength={500} />
        </label>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <button disabled={busy}>Apply placement</button>
      </form>
    </details>
  );
}
