"use client";
import {
  cloneElement,
  useId,
  useRef,
  useCallback,
  useEffect,
  useState,
} from "react";
import Link from "next/link";
import { AcademicStudio } from "./academic-studio";
import { AdmissionPanel } from "./admission-panel";
import { AccountPanel } from "./account-panel";
import { SettingsPanel } from "./settings-panel";
import { CampusAction, PlacementAction } from "./record-actions";
import { apiClient, type components } from "../../../clients/http";
type Me = components["schemas"]["Me"];
type Context = components["schemas"]["Workspace"];
type Campus = components["schemas"]["Campus"];
type Person = components["schemas"]["Person"];
type Course = components["schemas"]["Course"];
type Row = Record<string, string | number | boolean>;
type Pane =
  | "education"
  | "people"
  | "campuses"
  | "courses"
  | "enrollments"
  | "families"
  | "responsibilities"
  | "audit"
  | "deliveries"
  | "security"
  | "account"
  | "settings"
  | "intake";
const text = {
  en: {
    workspace: "Workspace",
    education: "Learning & assessment",
    people: "People",
    campuses: "Campuses",
    courses: "Courses",
    enrollments: "Admissions",
    families: "Family links",
    responsibilities: "Teaching responsibilities",
    audit: "Audit ledger",
    deliveries: "Delivery operations",
    security: "Access and security",
    intake: "Enrollment intake",
    account: "My account",
    settings: "Institution settings",
    institution: "Institution",
    campus: "Campus",
    all: "All permitted campuses",
    refresh: "Refresh",
    signout: "Sign out",
    save: "Save record",
    empty: "No records in this context.",
    loading: "Loading workspace…",
    name: "Name",
    email: "Email address",
    role: "Responsibility",
    invite: "Invite a person",
    search: "Search people",
    scope: "Access follows your current responsibilities.",
    create: "Create",
    saved: "Saved. The record and audit entry were committed.",
    retry: "Retry",
    reason: "Reason",
    state: "State",
    subject: "Subject",
    course: "Course",
    learner: "Learner",
    guardian: "Guardian",
    relationship: "Relationship",
    start: "Start date",
    end: "End date (optional)",
    code: "Code",
    address: "Address",
    year: "Academic year",
    neworg: "Add institution",
    timezone: "Timezone",
    locale: "Language",
    access: "Manage access",
    suspend: "Suspend",
    restore: "Restore",
    mfa: "Authenticator protection",
    begin: "Set up authenticator",
    confirm: "Confirm code",
    revoke: "Revoke session",
    active: "Your active sessions",
    local: "Development installation · synthetic data only",
  },
  hi: {
    education: "अध्ययन और मूल्यांकन",
    workspace: "कार्यस्थल",
    people: "लोग",
    campuses: "परिसर",
    courses: "पाठ्यक्रम",
    enrollments: "प्रवेश",
    families: "अभिभावक संबंध",
    responsibilities: "शिक्षण जिम्मेदारियाँ",
    audit: "लेखा परीक्षण",
    deliveries: "संदेश संचालन",
    security: "प्रवेश और सुरक्षा",
    intake: "प्रवेश अनुरोध",
    account: "मेरा खाता",
    settings: "संस्थान सेटिंग्स",
    institution: "संस्थान",
    campus: "परिसर",
    all: "सभी अनुमत परिसर",
    refresh: "ताज़ा करें",
    signout: "साइन आउट",
    save: "रिकॉर्ड सहेजें",
    empty: "इस संदर्भ में कोई रिकॉर्ड नहीं है।",
    loading: "कार्यस्थल लोड हो रहा है…",
    name: "नाम",
    email: "ईमेल पता",
    role: "जिम्मेदारी",
    invite: "व्यक्ति को आमंत्रित करें",
    search: "लोग खोजें",
    scope: "प्रवेश आपकी वर्तमान जिम्मेदारियों के अनुसार है।",
    create: "बनाएँ",
    saved: "रिकॉर्ड और परीक्षण प्रविष्टि सहेजे गए।",
    retry: "पुनः प्रयास",
    reason: "कारण",
    state: "स्थिति",
    subject: "विषय",
    course: "पाठ्यक्रम",
    learner: "शिक्षार्थी",
    guardian: "अभिभावक",
    relationship: "संबंध",
    start: "आरंभ तिथि",
    end: "समाप्ति तिथि (वैकल्पिक)",
    code: "कोड",
    address: "पता",
    year: "शैक्षणिक वर्ष",
    neworg: "संस्थान जोड़ें",
    timezone: "समय क्षेत्र",
    locale: "भाषा",
    access: "प्रवेश प्रबंधन",
    suspend: "निलंबित करें",
    restore: "पुनर्स्थापित करें",
    mfa: "प्रमाणक सुरक्षा",
    begin: "प्रमाणक सेट करें",
    confirm: "कोड की पुष्टि करें",
    revoke: "सत्र रद्द करें",
    active: "आपके सक्रिय सत्र",
    local: "विकास स्थापना · केवल काल्पनिक डेटा",
  },
};
const columns: Record<
  Exclude<Pane, "education" | "security" | "account" | "settings" | "intake">,
  string[]
> = {
  people: ["name", "email", "role", "state", "access_state"],
  campuses: ["name", "code", "address", "state"],
  courses: ["name", "academic_year"],
  enrollments: ["name", "reference", "course", "state"],
  families: ["guardian", "learner", "relationship", "state"],
  responsibilities: ["staff", "course", "subject", "starts_on", "ends_on"],
  audit: ["actor", "action", "reason", "time", "request_id"],
  deliveries: ["recipient", "kind", "state", "attempts", "error"],
};
function Label({
  name,
  children,
}: {
  name: string;
  children: React.ReactElement<{ id?: string }>;
}) {
  const id = useId();
  return (
    <div className="field">
      <label htmlFor={id}>{name}</label>
      {cloneElement(children, { id })}
    </div>
  );
}
function field(f: FormData, key: string) {
  return String(f.get(key) ?? "");
}
export function Workspace({
  initialMe,
  initialContexts,
}: {
  initialMe: Me;
  initialContexts: Context[];
}) {
  const [hydrated, setHydrated] = useState(false);
  useEffect(() => setHydrated(true), []);
  const [me, setMe] = useState<Me | null>(initialMe),
    [contexts, setContexts] = useState<Context[]>(initialContexts),
    [org, setOrg] = useState(initialContexts[0]?.id ?? ""),
    [campus, setCampus] = useState(""),
    [campuses, setCampuses] = useState<Campus[]>([]),
    [people, setPeople] = useState<Person[]>([]),
    [courses, setCourses] = useState<Course[]>([]),
    [pane, setPane] = useState<Pane>(
      initialContexts.some((x) => ["owner", "coordinator"].includes(x.role))
        ? "people"
        : "account",
    ),
    [dataset, setDataset] = useState<{ key: string; rows: Row[] }>({
      key: "",
      rows: [],
    }),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [loading, setLoading] = useState(true),
    [busy, setBusy] = useState(false),
    [query, setQuery] = useState(""),
    [language, setLanguage] = useState<"en" | "hi">("en"),
    [theme, setTheme] = useState("system"),
    [large, setLarge] = useState(false),
    [mfa, setMfa] = useState<{ secret: string; url: string } | null>(null),
    [codes, setCodes] = useState<string[]>([]),
    [sessions, setSessions] = useState<components["schemas"]["Session"][]>([]);
  const pendingLoad = useRef<AbortController | null>(null);
  const datasetKey = [org, campus, pane, query].join("|");
  const rows = dataset.key === datasetKey ? dataset.rows : [];
  function setRows(rows: Row[]) {
    setDataset({ key: datasetKey, rows });
  }
  const memberScope = contexts.find((x) => x.id === org)?.campus_id ?? "";
  const t = text[language];
  const context = contexts.find((x) => x.id === org);
  const owner = contexts.some((x) => x.id === org && x.role === "owner");
  const manager =
    owner || contexts.some((x) => x.id === org && x.role === "coordinator");
  const api = useCallback(() => apiClient(me?.csrf), [me?.csrf]);
  const formatTime = (value: string) =>
    new Intl.DateTimeFormat(language === "hi" ? "hi-IN" : "en-IN", {
      dateStyle: "medium",
      timeStyle: "short",
      timeZone: context?.timezone ?? "UTC",
    }).format(new Date(value));
  useEffect(() => {
    const saved = localStorage.getItem("workspace-language");
    if (saved === "hi") setLanguage(saved);
    setLarge(localStorage.getItem("workspace-large") === "true");
    const th = localStorage.getItem("workspace-theme") ?? "system";
    setTheme(th);
  }, []);
  useEffect(() => {
    document.documentElement.lang = language;
    document.documentElement.dataset.theme = theme;
    document.documentElement.dataset.large = String(large);
    localStorage.setItem("workspace-language", language);
    localStorage.setItem("workspace-theme", theme);
    localStorage.setItem("workspace-large", String(large));
  }, [language, theme, large]);
  const loadContexts = useCallback(async () => {
    const result = await api().GET("/api/v1/workspaces");
    if (result.error) throw new Error(result.error.message);
    setContexts(result.data);
    return result.data;
  }, [api]);
  useEffect(() => {
    const last = localStorage.getItem("workspace-institution");
    setOrg(
      initialContexts.find((x) => x.id === last)?.id ??
        initialContexts[0]?.id ??
        "",
    );
    if (!initialContexts.length) setLoading(false);
  }, [initialContexts]);
  useEffect(() => {
    if (!me || !org) return;
    let active = true;
    setCampus("");
    setRows([]);
    setCampuses([]);
    localStorage.setItem("workspace-institution", org);
    api()
      .GET("/api/v1/institutions/{org}/campuses", { params: { path: { org } } })
      .then((r) => {
        if (!active) return;
        if (r.error) setError(r.error.message);
        else {
          setCampuses(r.data);
          const scoped = memberScope;
          if (scoped) setCampus(scoped);
        }
      })
      .catch(() => setError("Connection unavailable."));
    return () => {
      active = false;
    };
  }, [org, me?.id, api, memberScope]);
  const load = useCallback(async () => {
    if (!me || !org) return;
    pendingLoad.current?.abort();
    const controller = new AbortController();
    pendingLoad.current = controller;
    const signal = controller.signal;
    setLoading(true);
    setError("");
    setRows([]);
    try {
      const params = { path: { org }, query: { campus, q: query } };
      let result;
      switch (pane) {
        case "people":
          result = await api().GET("/api/v1/institutions/{org}/people", {
            params,
            signal,
          });
          if (result.data) setPeople(result.data);
          break;
        case "campuses":
          result = await api().GET("/api/v1/institutions/{org}/campuses", {
            params,
            signal,
          });
          if (result.data) setCampuses(result.data);
          break;
        case "courses":
          result = await api().GET("/api/v1/institutions/{org}/courses", {
            params,
            signal,
          });
          if (result.data) setCourses(result.data);
          break;
        case "enrollments":
          result = await api().GET("/api/v1/institutions/{org}/enrollments", {
            params,
            signal,
          });
          break;
        case "families":
          result = await api().GET("/api/v1/institutions/{org}/families", {
            params,
            signal,
          });
          break;
        case "responsibilities":
          result = await api().GET(
            "/api/v1/institutions/{org}/responsibilities",
            { params, signal },
          );
          break;
        case "audit":
          result = await api().GET("/api/v1/institutions/{org}/audit", {
            params,
            signal,
          });
          break;
        case "deliveries":
          result = await api().GET("/api/v1/institutions/{org}/deliveries", {
            params,
            signal,
          });
          break;
        case "education":
        case "intake":
        case "account":
        case "settings":
          return;
        case "security": {
          const r = await api().GET("/api/v1/sessions", { signal });
          if (r.error) throw new Error(r.error.message);
          setSessions(r.data);
          return;
        }
      }
      if (result.error) throw new Error(result.error.message);
      setRows(result.data as Row[]);
      if (["enrollments", "families", "responsibilities"].includes(pane)) {
        const [p, c] = await Promise.all([
          api().GET("/api/v1/institutions/{org}/people", { params, signal }),
          api().GET("/api/v1/institutions/{org}/courses", { params, signal }),
        ]);
        if (p.data) setPeople(p.data);
        if (c.data) setCourses(c.data);
        if (p.error || c.error)
          throw new Error(p.error?.message ?? c.error?.message);
      }
    } catch (e) {
      if (signal.aborted) return;
      setError(
        e instanceof Error ? e.message : "Connection failed. Retry to reload.",
      );
    } finally {
      if (!signal.aborted) setLoading(false);
    }
  }, [api, me, org, campus, pane, query]);
  useEffect(() => {
    void load();
    return () => pendingLoad.current?.abort();
  }, [load]);
  async function act(
    operation: () => Promise<{ error?: { message: string }; data?: unknown }>,
    after?: () => void,
  ) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const r = await operation();
      if (r.error) throw new Error(r.error.message);
      setNotice(t.saved);
      after?.();
      await load();
      return r.data;
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Connection failed. Your input is retained.",
      );
    } finally {
      setBusy(false);
    }
  }
  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = e.currentTarget;
    const f = new FormData(form);
    const path = { org };
    const reset = () => form.reset();
    switch (pane) {
      case "campuses":
        await act(
          () =>
            api().POST("/api/v1/institutions/{org}/campuses", {
              params: { path, header: { "X-CSRF-Token": me?.csrf ?? "" } },
              body: {
                name: field(f, "name"),
                code: field(f, "code"),
                address: field(f, "address"),
              },
            }),
          reset,
        );
        break;
      case "people":
        await act(
          () =>
            api().POST("/api/v1/institutions/{org}/invitations", {
              params: { path, header: { "X-CSRF-Token": me?.csrf ?? "" } },
              body: {
                name: field(f, "name"),
                email: field(f, "email"),
                role: field(
                  f,
                  "role",
                ) as components["schemas"]["Invite"]["role"],
                campus_id: field(f, "campus_id"),
              },
            }),
          reset,
        );
        break;
      case "courses":
        await act(
          () =>
            api().POST("/api/v1/institutions/{org}/courses", {
              params: { path, header: { "X-CSRF-Token": me?.csrf ?? "" } },
              body: {
                name: field(f, "name"),
                campus_id: field(f, "campus_id"),
                academic_year: field(f, "academic_year"),
              },
            }),
          reset,
        );
        break;
      case "enrollments":
        await act(
          () =>
            api().POST("/api/v1/institutions/{org}/enrollments", {
              params: { path, header: { "X-CSRF-Token": me?.csrf ?? "" } },
              body: {
                person_id: field(f, "person_id"),
                campus_id: field(f, "campus_id"),
                course_id: field(f, "course_id"),
                start_date: field(f, "start_date"),
              },
            }),
          reset,
        );
        break;
      case "families":
        await act(
          () =>
            api().POST("/api/v1/institutions/{org}/families", {
              params: { path, header: { "X-CSRF-Token": me?.csrf ?? "" } },
              body: {
                guardian_id: field(f, "guardian_id"),
                learner_id: field(f, "learner_id"),
                campus_id: field(f, "campus_id"),
                relationship: field(f, "relationship"),
              },
            }),
          reset,
        );
        break;
      case "responsibilities":
        await act(
          () =>
            api().POST("/api/v1/institutions/{org}/responsibilities", {
              params: { path, header: { "X-CSRF-Token": me?.csrf ?? "" } },
              body: {
                staff_id: field(f, "staff_id"),
                campus_id: field(f, "campus_id"),
                course_id: field(f, "course_id"),
                subject: field(f, "subject"),
                starts_on: field(f, "starts_on"),
                ends_on: field(f, "ends_on"),
              },
            }),
          reset,
        );
        break;
    }
  }
  if (!me)
    return (
      <main id="main" className="public">
        <h1>{t.workspace}</h1>
        {error ? (
          <>
            <p role="alert">{error}</p>
            <button onClick={() => location.reload()}>{t.retry}</button>
          </>
        ) : (
          <p role="status">{t.loading}</p>
        )}
      </main>
    );
  return (
    <fieldset className="hydration-guard" disabled={!hydrated}>
      <legend className="sr-only">Workspace tools</legend>
      <div className="workspace">
        <aside className="rail">
          <Link href="/" className="wordmark">
            LW /<br />
            Learning Workspace
          </Link>
          <p className="rail-caption">{t.workspace}</p>
          <nav aria-label={t.workspace}>
            {(
              [
                "education",
                "people",
                "intake",
                "enrollments",
                "families",
                "responsibilities",
                "courses",
                "campuses",
                "audit",
                "deliveries",
                "security",
                "account",
                "settings",
              ] as Pane[]
            )
              .filter((p) =>
                p === "settings"
                  ? owner
                  : manager ||
                    p === "education" ||
                    p === "security" ||
                    p === "account",
              )
              .map((p) => (
                <button
                  key={p}
                  aria-current={pane === p ? "page" : undefined}
                  onClick={() => {
                    setPane(p);
                    setNotice("");
                    setQuery("");
                  }}
                >
                  {t[p]}
                </button>
              ))}
          </nav>
          <div className="rail-bottom">
            <p>
              {me.name}
              <br />
              <span>{me.email}</span>
            </p>
            <button
              onClick={() =>
                void act(
                  () =>
                    api().POST("/api/v1/sign-out", {
                      params: { header: { "X-CSRF-Token": me.csrf } },
                    }),
                  () => location.assign("/sign-in"),
                )
              }
            >
              {t.signout}
            </button>
          </div>
        </aside>
        <div className="workbody">
          <header className="context-bar">
            <Label name={t.institution}>
              <select
                value={org}
                onChange={(e) => {
                  setOrg(e.target.value);
                  setError("");
                }}
              >
                {Array.from(
                  new Map(contexts.map((c) => [c.id, c])).values(),
                ).map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Label>
            <Label name={t.campus}>
              <select
                value={campus}
                onChange={(e) => {
                  setPeople([]);
                  setCourses([]);
                  setCampus(e.target.value);
                }}
              >
                <option value="">{t.all}</option>
                {campuses.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Label>
            <Label name={t.locale}>
              <select
                value={language}
                onChange={(e) => setLanguage(e.target.value as "en" | "hi")}
              >
                <option value="en">English</option>
                <option value="hi">हिन्दी</option>
              </select>
            </Label>
            <Label name="Appearance">
              <select value={theme} onChange={(e) => setTheme(e.target.value)}>
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </Label>
            <label className="check">
              <input
                type="checkbox"
                checked={large}
                onChange={(e) => setLarge(e.target.checked)}
              />
              A+
            </label>
          </header>
          <main id="main" className="workmain">
            <div className="page-heading">
              <div>
                <p className="eyebrow">
                  {context?.name} / {context?.timezone}
                </p>
                <h1>{t[pane]}</h1>
                <p className="muted">{t.scope}</p>
              </div>
              <button onClick={() => void load()} disabled={loading}>
                {t.refresh}
              </button>
            </div>
            {me.environment !== "production" && (
              <p className="environment">{t.local}</p>
            )}
            {error && (
              <div className="error" role="alert">
                {error} <button onClick={() => void load()}>{t.retry}</button>
              </div>
            )}
            {notice && (
              <p role="status" className="notice">
                {notice}
              </p>
            )}
            {pane === "people" && (
              <form
                className="search"
                onSubmit={(e) => {
                  e.preventDefault();
                  void load();
                }}
              >
                <label>
                  {t.search}
                  <input
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    maxLength={100}
                  />
                </label>
              </form>
            )}
            {pane !== "education" &&
              pane !== "security" &&
              pane !== "account" &&
              pane !== "settings" &&
              pane !== "intake" && (
                <>
                  <section
                    className="ledger"
                    aria-label={t[pane]}
                    aria-busy={loading}
                  >
                    {loading ? (
                      <p role="status">{t.loading}</p>
                    ) : rows.length === 0 ? (
                      <p className="empty">{t.empty}</p>
                    ) : (
                      <div
                        className="table-scroll"
                        tabIndex={0}
                        role="region"
                        aria-label={`${t[pane]} · scroll to see all columns`}
                      >
                        <table>
                          <caption>
                            {t[pane]} · {rows.length} records (bounded view)
                          </caption>
                          <thead>
                            <tr>
                              {columns[pane].map((c) => (
                                <th key={c}>{c.replaceAll("_", " ")}</th>
                              ))}
                              {((pane === "people" && owner) ||
                                (pane === "families" && owner) ||
                                (pane === "deliveries" && owner) ||
                                (pane === "campuses" && manager) ||
                                (pane === "enrollments" && manager)) && (
                                <th>{t.access}</th>
                              )}
                            </tr>
                          </thead>
                          <tbody>
                            {rows.map((row, i) => (
                              <tr key={String(row.id ?? i)}>
                                {columns[pane].map((c) => (
                                  <td key={c}>
                                    {c === "time"
                                      ? formatTime(String(row[c]))
                                      : String(row[c] ?? "—")}
                                  </td>
                                ))}
                                {pane === "campuses" && manager && (
                                  <td>
                                    <CampusAction
                                      org={org}
                                      csrf={me.csrf}
                                      row={row as Campus}
                                      onSaved={() => void load()}
                                    />
                                  </td>
                                )}
                                {pane === "enrollments" && manager && (
                                  <td>
                                    <PlacementAction
                                      org={org}
                                      csrf={me.csrf}
                                      row={
                                        row as components["schemas"]["Enrollment"]
                                      }
                                      campuses={campuses}
                                      onSaved={() => void load()}
                                    />
                                  </td>
                                )}
                                {pane === "people" && owner && (
                                  <td>
                                    {row.role !== "owner" &&
                                      row.id !== me.id && (
                                        <MembershipAction
                                          person={row}
                                          label={
                                            row.state === "active"
                                              ? t.suspend
                                              : t.restore
                                          }
                                          reasonLabel={t.reason}
                                          disabled={busy}
                                          onSubmit={(reason) =>
                                            act(() =>
                                              api().PATCH(
                                                "/api/v1/institutions/{org}/memberships/{id}",
                                                {
                                                  params: {
                                                    path: {
                                                      org,
                                                      id: String(
                                                        row.membership_id,
                                                      ),
                                                    },
                                                    header: {
                                                      "X-CSRF-Token": me.csrf,
                                                    },
                                                  },
                                                  body: {
                                                    state:
                                                      row.state === "active"
                                                        ? "suspended"
                                                        : "active",
                                                    version: Number(
                                                      row.version,
                                                    ),
                                                    reason,
                                                  },
                                                },
                                              ),
                                            )
                                          }
                                        />
                                      )}
                                  </td>
                                )}
                                {pane === "families" && owner && (
                                  <td>
                                    {row.state === "approved" && (
                                      <button
                                        disabled={busy}
                                        onClick={() => {
                                          if (
                                            confirm(
                                              "Remove this family link and end its access?",
                                            )
                                          )
                                            void act(() =>
                                              api().DELETE(
                                                "/api/v1/institutions/{org}/families/{id}",
                                                {
                                                  params: {
                                                    path: {
                                                      org,
                                                      id: String(row.id),
                                                    },
                                                    header: {
                                                      "X-CSRF-Token": me.csrf,
                                                    },
                                                  },
                                                },
                                              ),
                                            );
                                        }}
                                      >
                                        Remove link
                                      </button>
                                    )}
                                  </td>
                                )}
                                {pane === "deliveries" && owner && (
                                  <td>
                                    {row.state === "failed" && (
                                      <button
                                        disabled={busy}
                                        onClick={() =>
                                          void act(() =>
                                            api().POST(
                                              "/api/v1/institutions/{org}/deliveries/{id}/retry",
                                              {
                                                params: {
                                                  path: {
                                                    org,
                                                    id: String(row.id),
                                                  },
                                                  header: {
                                                    "X-CSRF-Token": me.csrf,
                                                  },
                                                },
                                              },
                                            ),
                                          )
                                        }
                                      >
                                        {t.retry}
                                      </button>
                                    )}
                                  </td>
                                )}
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </section>
                  {manager &&
                    [
                      "people",
                      "campuses",
                      "courses",
                      "enrollments",
                      "families",
                      "responsibilities",
                    ].includes(pane) && (
                      <section className="editor">
                        <h2>
                          {pane === "people"
                            ? t.invite
                            : t.create + " · " + t[pane]}
                        </h2>
                        <form
                          onSubmit={submit}
                          className="field-grid"
                          aria-busy={busy}
                        >
                          {["people", "campuses", "courses"].includes(pane) && (
                            <Label name={t.name}>
                              <input
                                name="name"
                                required
                                minLength={2}
                                maxLength={120}
                              />
                            </Label>
                          )}
                          {pane === "people" && (
                            <>
                              <Label name={t.email}>
                                <input
                                  type="email"
                                  name="email"
                                  required
                                  maxLength={254}
                                />
                              </Label>
                              <Label name={t.role}>
                                <select name="role">
                                  <option value="learner">Learner</option>
                                  <option value="guardian">Guardian</option>
                                  <option value="instructor">Instructor</option>
                                  <option value="reviewer">Reviewer</option>
                                  {owner && (
                                    <option value="coordinator">
                                      Coordinator
                                    </option>
                                  )}
                                </select>
                              </Label>
                            </>
                          )}
                          {pane !== "campuses" && (
                            <CampusSelect
                              campuses={campuses}
                              value={campus}
                              label={t.campus}
                            />
                          )}
                          {pane === "campuses" && (
                            <>
                              <Label name={t.code}>
                                <input name="code" required maxLength={20} />
                              </Label>
                              <Label name={t.address}>
                                <input name="address" maxLength={500} />
                              </Label>
                            </>
                          )}
                          {pane === "courses" && (
                            <Label name={t.year}>
                              <input
                                name="academic_year"
                                required
                                minLength={4}
                                maxLength={20}
                              />
                            </Label>
                          )}
                          {pane === "enrollments" && (
                            <>
                              <PersonSelect
                                people={people}
                                role="learner"
                                name="person_id"
                                label={t.learner}
                              />
                              <CourseSelect
                                courses={courses}
                                label={t.course}
                              />
                              <Label name={t.start}>
                                <input type="date" name="start_date" required />
                              </Label>
                            </>
                          )}
                          {pane === "families" && (
                            <>
                              <PersonSelect
                                people={people}
                                role="guardian"
                                name="guardian_id"
                                label={t.guardian}
                              />
                              <PersonSelect
                                people={people}
                                role="learner"
                                name="learner_id"
                                label={t.learner}
                              />
                              <Label name={t.relationship}>
                                <input
                                  name="relationship"
                                  required
                                  minLength={2}
                                  maxLength={80}
                                />
                              </Label>
                            </>
                          )}
                          {pane === "responsibilities" && (
                            <>
                              <PersonSelect
                                people={people}
                                role="instructor"
                                name="staff_id"
                                label={t.role}
                              />
                              <CourseSelect
                                courses={courses}
                                label={t.course}
                              />
                              <Label name={t.subject}>
                                <input
                                  name="subject"
                                  required
                                  maxLength={100}
                                />
                              </Label>
                              <Label name={t.start}>
                                <input name="starts_on" type="date" required />
                              </Label>
                              <Label name={t.end}>
                                <input name="ends_on" type="date" />
                              </Label>
                            </>
                          )}
                          <div className="form-actions">
                            <button className="primary" disabled={busy}>
                              {busy ? "…" : t.save}
                            </button>
                          </div>
                        </form>
                      </section>
                    )}
                  {pane === "campuses" && owner && (
                    <section className="editor">
                      <h2>{t.neworg}</h2>
                      <form
                        className="field-grid"
                        onSubmit={(e) => {
                          e.preventDefault();
                          const form = e.currentTarget,
                            f = new FormData(form);
                          void act(
                            () =>
                              api().POST("/api/v1/institutions", {
                                params: { header: { "X-CSRF-Token": me.csrf } },
                                body: {
                                  name: field(f, "name"),
                                  timezone: field(f, "timezone"),
                                  locale: field(f, "locale") as "en" | "hi",
                                },
                              }),
                            () => {
                              form.reset();
                              void loadContexts();
                            },
                          );
                        }}
                      >
                        <Label name={t.name}>
                          <input
                            name="name"
                            required
                            minLength={2}
                            maxLength={120}
                          />
                        </Label>
                        <Label name={t.timezone}>
                          <input
                            name="timezone"
                            defaultValue="Asia/Kolkata"
                            required
                          />
                        </Label>
                        <Label name={t.locale}>
                          <select name="locale">
                            <option value="en">English</option>
                            <option value="hi">हिन्दी</option>
                          </select>
                        </Label>
                        <button className="primary" disabled={busy}>
                          {t.create}
                        </button>
                      </form>
                    </section>
                  )}
                </>
              )}
            {pane === "security" && (
              <>
                <section className="editor">
                  <h2>{t.mfa}</h2>
                  {me.mfa_enabled ? (
                    <p>Enabled. Keep your recovery codes in a safe place.</p>
                  ) : (
                    <>
                      <p>
                        Use an authenticator app. Confirm a code before
                        protection is enabled.
                      </p>
                      <button
                        disabled={busy}
                        onClick={() =>
                          void act(() =>
                            api().POST("/api/v1/mfa/enroll", {
                              params: { header: { "X-CSRF-Token": me.csrf } },
                            }),
                          ).then((data) => {
                            if (data)
                              setMfa(data as { secret: string; url: string });
                          })
                        }
                      >
                        {t.begin}
                      </button>
                    </>
                  )}
                  {mfa && !me.mfa_enabled && (
                    <>
                      <p>Enter this secret in your authenticator:</p>
                      <code className="secret">{mfa.secret}</code>
                      <form
                        onSubmit={(e) => {
                          e.preventDefault();
                          const f = new FormData(e.currentTarget);
                          void act(() =>
                            api().POST("/api/v1/mfa/confirm", {
                              params: { header: { "X-CSRF-Token": me.csrf } },
                              body: { code: field(f, "code") },
                            }),
                          ).then((data) => {
                            if (data) {
                              setCodes(
                                (data as { recovery_codes: string[] })
                                  .recovery_codes,
                              );
                              setMe({ ...me, mfa_enabled: true });
                              setMfa(null);
                            }
                          });
                        }}
                      >
                        <Label name={t.code}>
                          <input
                            name="code"
                            autoComplete="one-time-code"
                            inputMode="numeric"
                            required
                            pattern="[0-9]{6}"
                          />
                        </Label>
                        <button disabled={busy}>{t.confirm}</button>
                      </form>
                    </>
                  )}
                  {codes.length > 0 && (
                    <div className="notice">
                      <h3>Save your single-use recovery codes</h3>
                      <pre>{codes.join("\n")}</pre>
                      <button onClick={() => setCodes([])}>
                        I saved these codes
                      </button>
                    </div>
                  )}
                </section>
                <section className="ledger">
                  <h2>{t.active}</h2>
                  {sessions.map((x) => (
                    <article className="session" key={x.id}>
                      <div>
                        <strong>
                          {x.current ? "Current session" : x.device}
                        </strong>
                        <p>
                          {formatTime(x.last_seen)} · {x.device}
                        </p>
                      </div>
                      <button
                        disabled={busy}
                        onClick={() => {
                          if (confirm("Revoke this session?"))
                            void act(
                              () =>
                                api().DELETE("/api/v1/sessions/{id}", {
                                  params: {
                                    path: { id: x.id },
                                    header: { "X-CSRF-Token": me.csrf },
                                  },
                                }),
                              () => {
                                if (x.current) location.assign("/sign-in");
                              },
                            );
                        }}
                      >
                        {t.revoke}
                      </button>
                    </article>
                  ))}
                </section>
              </>
            )}
            {pane === "account" && (
              <AccountPanel
                csrf={me.csrf}
                onChanged={(profile) =>
                  setMe({ ...me, name: profile.name, email: profile.email })
                }
              />
            )}
            {pane === "settings" && (
              <SettingsPanel
                key={org}
                org={org}
                csrf={me.csrf}
                onSaved={() => void loadContexts()}
              />
            )}
            {pane === "education" && (
              <AcademicStudio
                key={org + campus}
                org={org}
                campus={campus}
                csrf={me.csrf}
                staff={contexts.some(
                  (c) =>
                    c.id === org &&
                    ["owner", "coordinator", "instructor"].includes(c.role),
                )}
              />
            )}
            {pane === "intake" && (
              <AdmissionPanel
                key={org + campus}
                org={org}
                campus={campus}
                csrf={me.csrf}
                campuses={campuses}
                timezone={context?.timezone ?? "UTC"}
              />
            )}
          </main>
        </div>
      </div>
    </fieldset>
  );
}
function MembershipAction({
  person,
  label,
  reasonLabel,
  disabled,
  onSubmit,
}: {
  person: Row;
  label: string;
  reasonLabel: string;
  disabled: boolean;
  onSubmit: (reason: string) => Promise<unknown>;
}) {
  const [open, setOpen] = useState(false);
  return open ? (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        const f = new FormData(e.currentTarget);
        void onSubmit(field(f, "reason")).then(() => setOpen(false));
      }}
    >
      <label>
        {reasonLabel}
        <input name="reason" required minLength={5} maxLength={500} />
      </label>
      <button disabled={disabled}>
        {label} {String(person.name)}
      </button>
      <button type="button" onClick={() => setOpen(false)}>
        Cancel
      </button>
    </form>
  ) : (
    <button disabled={disabled} onClick={() => setOpen(true)}>
      {label}
    </button>
  );
}

function CampusSelect({
  campuses,
  value,
  label,
}: {
  campuses: Campus[];
  value: string;
  label: string;
}) {
  return (
    <Label name={label}>
      <select name="campus_id" required defaultValue={value}>
        <option value="">—</option>
        {campuses.map((c) => (
          <option key={c.id} value={c.id}>
            {c.name}
          </option>
        ))}
      </select>
    </Label>
  );
}
function PersonSelect({
  people,
  role,
  name,
  label,
}: {
  people: Person[];
  role: string;
  name: string;
  label: string;
}) {
  return (
    <Label name={label}>
      <select name={name} required>
        <option value="">—</option>
        {people
          .filter((p) => p.role === role && p.state === "active")
          .map((p) => (
            <option key={p.membership_id} value={p.id}>
              {p.name} · {p.email}
            </option>
          ))}
      </select>
    </Label>
  );
}
function CourseSelect({
  courses,
  label,
}: {
  courses: Course[];
  label: string;
}) {
  return (
    <Label name={label}>
      <select name="course_id" required>
        <option value="">—</option>
        {courses.map((c) => (
          <option key={c.id} value={c.id}>
            {c.name}
          </option>
        ))}
      </select>
    </Label>
  );
}
