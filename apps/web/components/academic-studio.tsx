"use client";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
} from "react";
import {
  educationClient,
  type EducationClient,
  type A,
} from "../../../clients/academic";
import {
  LearningHome,
  GenerationFlow,
  SamplesFlow,
  ResultsFlow,
} from "./academic-flows";
import { type LearningView } from "./learning-navigation";
import { QuestionEditor } from "./question-editor";
import { AssessmentWorkspace, FacultyStudents } from "./assessment-workspace";
import { TaskProgress } from "./task-progress";
import { GridInspection } from "./grid-inspection";
import { OMRCheckerReader } from "./omrchecker-reader";
import { ImportChoiceRow, KeyImportPanel, MarkingRules } from "./import-review";
import { FormField as Field } from "./form-field";
type Q = A<"AcademicQuestion">;
type Paper = A<"AcademicPaperRecord">;
type Sitting = A<"AcademicSitting">;
type Answer = A<"AcademicAnswer">;
type Topic = A<"AcademicTopic">;
const blank = (): Answer => ({
  selection: [],
  number: "",
  text: "",
  visited: true,
  review: false,
  order: 0,
  invalid: false,
});
function values(e: FormEvent<HTMLFormElement>) {
  e.preventDefault();
  return new FormData(e.currentTarget);
}
const str = (f: FormData, k: string) => String(f.get(k) ?? "");
const nowLocal = () =>
  new Date(Date.now() - new Date().getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16);
function emptyQuestion(topic: string): Q {
  return {
    topic_id: topic,
    type: "single",
    stem: "",
    options: [
      { id: "A", text: "" },
      { id: "B", text: "" },
      { id: "C", text: "" },
      { id: "D", text: "" },
    ],
    key: [],
    rubric: [],
    solution: "",
    marks: "1",
    penalty: "0",
    difficulty: "standard",
    language: "en",
    provenance: "Original teacher-authored material",
    license: "Institution-owned",
  };
}
export function AcademicStudio({
  org,
  campus,
  csrf,
  staff,
  view,
  onNavigate,
  refreshToken,
}: {
  refreshToken: number;
  view: LearningView;
  onNavigate: (view: LearningView) => void;
  org: string;
  campus: string;
  csrf: string;
  staff: boolean;
}) {
  const api = useMemo(() => educationClient(org, csrf), [org, csrf]);
  const tab = view;
  const setTab = (next: LearningView) => {
    setSitting(null);
    onNavigate(next);
  };
  const [bankQuery, setBankQuery] = useState("");
  const [bankFilter, setBankFilter] = useState("all");
  const [topics, setTopics] = useState<Topic[]>([]);
  const [courses, setCourses] = useState<A<"AcademicCourse">[]>([]);
  const [topic, setTopic] = useState("");
  const topicRef = useRef(topic);
  topicRef.current = topic;
  const [bank, setBank] = useState<Q[]>([]);
  const [papers, setPapers] = useState<Paper[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [paper, setPaperState] = useState<Paper | null>(null);
  const setPaper = (p: Paper | null) => {
    setPaperState(p);
    if (p) {
      const [route, query] = window.location.hash.split("?");
      const params = new URLSearchParams(query);
      params.set("paper", p.id);
      window.location.hash = route + "?" + params;
    }
  };
  useEffect(() => {
    const restore = () => {
      const params = new URLSearchParams(window.location.hash.split("?")[1]);
      const id = params.get("paper");
      const found = papers.find((p) => p.id === id);
      if (found) setPaperState(found);
    };
    restore();
    window.addEventListener("hashchange", restore);
    return () => window.removeEventListener("hashchange", restore);
  }, [papers]);
  const [imports, setImports] = useState<A<"AcademicImport">[]>([]);
  const [uploadingPaper, setUploadingPaper] = useState(false);
  const [importPollError, setImportPollError] = useState("");
  const importReviewRef = useRef<HTMLElement>(null);
  useEffect(() => {
    if (uploadingPaper && tab === "import")
      document
        .getElementById("paper-processing")
        ?.scrollIntoView({ block: "start", behavior: "instant" });
  }, [uploadingPaper, tab]);
  const [preview, setPreviewState] = useState<A<"AcademicImport"> | null>(null);
  const previewRef = useRef(preview);
  previewRef.current = preview;
  const setPreview = (next: A<"AcademicImport"> | null) => {
    setPreviewState(next);
    if (tab !== "import") return;
    const [route, query] = window.location.hash.split("?");
    const params = new URLSearchParams(query);
    if (next) {
      params.set("import", next.id);
      params.set("topic", next.topic_id);
    } else {
      params.delete("import");
      params.delete("topic");
    }
    window.location.hash = route + (params.size ? "?" + params : "");
  };
  useEffect(() => {
    if (tab !== "import" || !staff) return;
    let active = true;
    const restore = () => {
      const params = new URLSearchParams(window.location.hash.split("?")[1]);
      const id = params.get("import");
      if (!id || id === previewRef.current?.id) return;
      void api
        .detail("imports", id)
        .then((next) => {
          if (active) {
            setTopic(next.topic_id);
            setPreviewState(next);
          }
        })
        .catch((e) => {
          if (active) setError(e.message);
        });
    };
    restore();
    window.addEventListener("hashchange", restore);
    return () => {
      active = false;
      window.removeEventListener("hashchange", restore);
    };
  }, [api, staff, tab]);
  useEffect(() => {
    if (!preview || preview.state !== "extracting") return;
    let active = true,
      pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        const [next, currentJobs] = await Promise.all([
          api.detail("imports", preview.id),
          api.list("jobs", { campus }),
        ]);
        if (active) {
          setPreview(next);
          setJobs(currentJobs);
          setImportPollError("");
        }
      } catch {
        if (active)
          setImportPollError(
            "Connection interrupted. Your upload is saved. We are trying to reconnect; you can also reload its status.",
          );
      } finally {
        pending = false;
      }
    }, 2000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, campus, preview?.id, preview?.state]);
  const [edit, setEdit] = useState<Q | null>(null);
  const editorRef = useRef<HTMLElement>(null);
  useEffect(() => {
    if (edit) {
      editorRef.current?.scrollIntoView({
        block: "start",
        behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches
          ? "instant"
          : "smooth",
      });
      editorRef.current
        ?.querySelector("textarea")
        ?.focus({ preventScroll: true });
    }
  }, [edit]);
  const [importIndex, setImportIndex] = useState<number | null>(null);
  const [jobs, setJobs] = useState<A<"AcademicJob">[]>([]);
  const [generationJob, setGenerationJob] = useState<A<"AcademicJob">>();
  const recordJob = useCallback((next: A<"AcademicJob">) => {
    setGenerationJob(next);
    setJobs((current) => [next, ...current.filter((j) => j.id !== next.id)]);
  }, []);
  useEffect(() => {
    if (tab !== "jobs" || !staff) return;
    let active = true,
      pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        const next = await api.list("jobs", { campus });
        if (active) setJobs(next);
      } catch {
        if (active)
          setError(
            "Task status could not refresh. Saved work is retained; reconnect or use Refresh tasks.",
          );
      } finally {
        pending = false;
      }
    }, 2500);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, campus, staff, tab]);
  const [assignments, setAssignments] = useState<A<"AcademicAssigned">[]>([]);
  const [sitting, setSitting] = useState<Sitting | null>(null);
  const [error, setError] = useState("");
  const errorRef = useRef<HTMLParagraphElement>(null);
  useEffect(() => {
    if (error) {
      errorRef.current?.scrollIntoView({ block: "center" });
      errorRef.current?.focus({ preventScroll: true });
    }
  }, [error]);
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [dataReady, setDataReady] = useState(false);
  const refresh = useCallback(async () => {
    if (staff) {
      const [t, c, p, j] = await Promise.all([
        api.list("topics", { campus }),
        api.list("courses", { campus }),
        api.list("papers", { campus }),
        api.list("jobs", { campus }),
      ]);
      setTopics(t);
      setCourses(c);
      setPapers(p);
      setJobs(j);
    } else setAssignments(await api.list("assignments"));
  }, [api, campus, staff]);
  useEffect(() => {
    let active = true;
    refresh()
      .then(() => {
        if (active) setDataReady(true);
      })
      .catch((e) => {
        if (active) setError(String(e.message));
      });
    return () => {
      active = false;
    };
  }, [refresh, refreshToken]);
  useEffect(() => {
    if (!topic || !staff) {
      setBank([]);
      setImports([]);
      return;
    }
    let active = true;
    Promise.all([
      api.list("questions", { topic }),
      api.list("imports", { topic }),
    ])
      .then(([b, i]) => {
        if (active) {
          setBank(b);
          setImports(i);
        }
      })
      .catch((e) => {
        if (active) setError(e.message);
      });
    return () => {
      active = false;
    };
  }, [api, topic, staff, refreshToken]);
  const run = async (fn: () => Promise<unknown>, message = "Saved.") => {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await fn();
      setNotice(message);
      await refresh();
      const activeTopic = topicRef.current;
      if (activeTopic && staff) {
        const [nextBank, nextImports] = await Promise.all([
          api.list("questions", { topic: activeTopic }),
          api.list("imports", { topic: activeTopic }),
        ]);
        if (topicRef.current === activeTopic) {
          setBank(nextBank);
          setImports(nextImports);
        }
      }
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Connection interrupted. Saved work is retained; retry.",
      );
    } finally {
      setBusy(false);
    }
  };
  useEffect(() => {
    if (!topic && topics.length) setTopic(topics[0].id);
  }, [topics, topic]);
  useEffect(() => {
    setSitting(null);
    setEdit(null);
    setImportIndex(null);
    setNotice("");
    setError("");
  }, [view]);
  const extractionJob = jobs.find(
    (j) =>
      j.kind === "extract_document" &&
      j.resource_id === preview?.id &&
      (j.input as Record<string, unknown>)?.mode !== "answer_key",
  );
  const emptyExtraction =
    preview?.state === "preview" && preview.rows.length === 0;
  const importState = uploadingPaper
    ? "uploading"
    : preview?.state === "extracting"
      ? ((extractionJob?.state === "completed"
          ? "running"
          : extractionJob?.state) ?? "queued")
      : emptyExtraction
        ? "failed"
        : preview
          ? "completed"
          : "";
  const importActive = uploadingPaper || preview?.state === "extracting";
  const chosen = topics.find((t) => t.id === topic);
  if (!dataReady)
    return (
      <section className="card" aria-busy={!error}>
        {error ? (
          <>
            <p className="error" role="alert">
              {error}
            </p>
            <button
              onClick={() => {
                setError("");
                void refresh()
                  .then(() => setDataReady(true))
                  .catch((e) => setError(e.message));
              }}
            >
              Retry loading workspace
            </button>
          </>
        ) : (
          <p role="status">
            Loading your {staff ? "teaching workspace" : "assessments"}…
          </p>
        )}
      </section>
    );
  if (sitting)
    return (
      <AssessmentWork
        key={sitting.id}
        api={api}
        initial={sitting}
        staff={staff}
        onBack={() => {
          setSitting(null);
          void run(refresh, "");
        }}
      />
    );
  return (
    <section
      className="academic"
      aria-label="Educational work"
      aria-busy={busy}
    >
      {error && (
        <p className="error" role="alert" ref={errorRef} tabIndex={-1}>
          {error}
        </p>
      )}
      {notice && (
        <p className="notice" role="status">
          {notice}
        </p>
      )}
      {staff ? (
        <>
          {(tab === "draft" || tab === "active") && (
            <AssessmentWorkspace
              key={tab}
              api={api}
              courses={courses}
              topics={topics}
              papers={papers}
              campus={campus}
              active={tab === "active"}
              onRefresh={refresh}
              onResults={(p) => {
                setPaper(p);
                setTab("results");
              }}
            />
          )}
          {tab === "students" && (
            <FacultyStudents api={api} courses={courses} />
          )}
          {tab === "home" && (
            <LearningHome
              openPaper={(p) => {
                setPaper(p);
                setTab("papers");
              }}
              navigate={setTab}
              topics={topics}
              papers={papers}
              jobs={jobs}
            />
          )}
          {tab === "generate" && (
            <GenerationFlow
              initialJob={
                generationJob ??
                jobs.find((j) => j.kind === "generate_questions")
              }
              onJob={recordJob}
              initialTopic={topic}
              api={api}
              topics={topics}
              courses={courses}
              onTopic={(id) => {
                setTopic(id);
                void refresh().catch((e) => setError(e.message));
              }}
              onReview={(id) =>
                void run(async () => {
                  setTopic(id);
                  setBank(await api.list("questions", { topic: id }));
                  setTab("curriculum");
                }, "")
              }
            />
          )}
          {tab === "samples" && (
            <SamplesFlow
              api={api}
              courses={courses}
              onOpen={(kit) =>
                void run(async () => {
                  setPaper(await api.detail("papers", kit.paper_id));
                  setTab("scan");
                }, "")
              }
            />
          )}
          {tab === "results" && (
            <ResultsFlow
              initialPaper={paper?.id ?? ""}
              api={api}
              papers={papers}
              onSitting={setSitting}
            />
          )}
          {(tab === "curriculum" || tab === "import" || tab === "author") && (
            <>
              {tab === "import" && (
                <>
                  <div className="actions">
                    <button onClick={() => setTab("curriculum")}>
                      ← Back to question bank
                    </button>
                  </div>
                  <ol
                    className="work-steps"
                    aria-label="Question upload workflow"
                  >
                    {[
                      "Upload paper",
                      "Extract questions",
                      "Review questions & key",
                      "Save to bank",
                    ].map((label, i) => {
                      const current = !preview
                        ? uploadingPaper
                          ? 1
                          : 0
                        : preview.state === "committed"
                          ? 3
                          : preview.state === "extracting"
                            ? 1
                            : 2;
                      return (
                        <li
                          key={label}
                          aria-current={current === i ? "step" : undefined}
                          className={i < current ? "is-done" : ""}
                        >
                          <span aria-hidden="true">
                            {i < current ? "✓" : i + 1}
                          </span>
                          {label}
                        </li>
                      );
                    })}
                  </ol>
                  {(uploadingPaper || preview) && (
                    <TaskProgress
                      id="paper-processing"
                      state={importState}
                      title={
                        uploadingPaper
                          ? "Uploading your question paper"
                          : preview?.state === "committed"
                            ? "Questions saved as drafts"
                            : preview?.state === "preview"
                              ? emptyExtraction
                                ? "No questions detected — review the extracted text"
                                : `${preview.rows.length} ${preview.rows.length === 1 ? "question" : "questions"} ready for review`
                              : importState === "queued"
                                ? "Paper saved — waiting to extract"
                                : importState === "running"
                                  ? "Extracting questions and options"
                                  : "Paper extraction needs attention"
                      }
                      detail={
                        uploadingPaper
                          ? "Keep this tab open until the upload is saved."
                          : emptyExtraction
                            ? "Your source is saved. Open the review below to correct the transcription or try another OCR engine."
                            : preview?.state === "preview"
                              ? "Check the wording and options, choose correct answers or apply a separate key, then save the reviewed questions to the bank."
                              : preview?.state === "committed"
                                ? "Open the question bank to approve these drafts before using them in an assessment."
                                : extractionJob?.error ||
                                  "Your original file is saved. Reading pages and arranging questions in order. You can leave this screen; extraction continues in Background tasks."
                      }
                      progress={extractionJob?.progress}
                    >
                      {emptyExtraction && (
                        <button
                          className="primary"
                          onClick={() =>
                            importReviewRef.current?.scrollIntoView({
                              block: "start",
                              behavior: "instant",
                            })
                          }
                        >
                          Fix extracted text
                        </button>
                      )}
                      {preview?.state === "preview" && !emptyExtraction && (
                        <button
                          className="primary"
                          onClick={() =>
                            importReviewRef.current?.scrollIntoView({
                              block: "start",
                              behavior: "instant",
                            })
                          }
                        >
                          Review extracted questions
                        </button>
                      )}
                      {preview?.state === "committed" && (
                        <button
                          className="primary"
                          onClick={() => setTab("curriculum")}
                        >
                          Open question bank
                        </button>
                      )}
                      {extractionJob &&
                        [
                          "failed",
                          "configuration_required",
                          "cancelled",
                        ].includes(extractionJob.state) &&
                        preview?.state === "extracting" && (
                          <button
                            disabled={busy}
                            onClick={() =>
                              void run(async () => {
                                await api.jobAction(extractionJob.id, "retry");
                                setJobs(await api.list("jobs", { campus }));
                              }, "")
                            }
                          >
                            Retry paper extraction
                          </button>
                        )}
                      <button onClick={() => setTab("jobs")}>
                        View background tasks
                      </button>
                    </TaskProgress>
                  )}
                  {importPollError && importActive && (
                    <p role="alert" className="error">
                      {importPollError}
                    </p>
                  )}
                </>
              )}
              <details
                className="upload-setup"
                open={tab !== "import" || !preview}
              >
                <summary>
                  {tab === "import"
                    ? preview
                      ? "Upload another paper or change topic"
                      : "1. Choose a topic and upload your paper"
                    : "Topic and learning outcome"}
                </summary>
                <div
                  className={
                    tab === "import" ? "academic-columns" : "topic-toolbar"
                  }
                >
                  <section className="card">
                    <h2>Choose a topic</h2>
                    <Field name="Topic">
                      <select
                        value={topic}
                        onChange={(e) => {
                          setTopic(e.target.value);
                          setPreview(null);
                          setEdit(null);
                          setSelected([]);
                        }}
                      >
                        <option value="">Select a topic</option>
                        {topics.map((t) => (
                          <option key={t.id} value={t.id}>
                            {t.subject} / {t.name}
                          </option>
                        ))}
                      </select>
                    </Field>
                    {chosen && <p>{chosen.outcome}</p>}
                    <details>
                      <summary>Add a topic</summary>
                      <form
                        onSubmit={(e) => {
                          const f = values(e);
                          void run(async () => {
                            const c = courses.find(
                              (c) => c.id === str(f, "course"),
                            );
                            if (!c)
                              throw new Error(
                                "Select a course. Create one in Courses if needed.",
                              );
                            const t = await api.topics({
                              id: "",
                              campus_id: c.campus_id,
                              course_id: c.id,
                              subject: str(f, "subject"),
                              name: str(f, "name"),
                              outcome: str(f, "outcome"),
                              language: str(f, "language"),
                            });
                            setTopic(t.id);
                          });
                        }}
                      >
                        <Field name="Course">
                          <select name="course" required>
                            <option value="">Select course</option>
                            {courses.map((c) => (
                              <option key={c.id} value={c.id}>
                                {c.name} / {c.academic_year}
                              </option>
                            ))}
                          </select>
                        </Field>
                        <Field name="Subject">
                          <input name="subject" required minLength={2} />
                        </Field>
                        <Field name="Topic name">
                          <input name="name" required minLength={2} />
                        </Field>
                        <Field name="Learning outcome">
                          <textarea name="outcome" required minLength={5} />
                        </Field>
                        <Field name="Content language">
                          <select name="language">
                            <option value="en">English</option>
                            <option value="hi">हिन्दी</option>
                          </select>
                        </Field>
                        <button disabled={busy}>Create topic</button>
                      </form>
                    </details>
                  </section>
                  {tab === "import" && (
                    <section className="card">
                      <h2>Upload your question paper</h2>
                      <p>
                        Use a PDF, photo, CSV, XLSX or JSON file. We will save a
                        preview first so you can check the questions and answers
                        before adding them to the bank.
                      </p>
                      <a href="/demo/questions.csv" download>
                        Download original mathematics CSV
                      </a>
                      <p className="actions">
                        {["physics", "chemistry", "biology"].map((subject) => (
                          <a
                            key={subject}
                            href={`/demo/${subject}.csv`}
                            download
                          >
                            {subject} example
                          </a>
                        ))}
                      </p>
                      <form
                        onSubmit={(e) => {
                          const f = values(e);
                          f.set("topic_id", topic);
                          void run(async () => {
                            setUploadingPaper(true);
                            setImportPollError("");
                            try {
                              setPreview(await api.uploadImport(f));
                            } finally {
                              setUploadingPaper(false);
                            }
                          }, "");
                        }}
                      >
                        <Field name="Question file">
                          <input
                            name="file"
                            type="file"
                            accept=".csv,.xlsx,.json,.pdf,.png,.jpg,.jpeg"
                            required
                          />
                        </Field>
                        <Field name="Extraction engine">
                          <select name="engine">
                            <option value="">Default engine</option>
                            <option value="lightonocr">
                              LightOnOCR-2-1B (local model)
                            </option>
                            <option value="tesseract">
                              Tesseract (local text OCR)
                            </option>
                          </select>
                        </Field>
                        <Field name="Source / provenance">
                          <input
                            name="provenance"
                            required
                            defaultValue="Original teacher material"
                          />
                        </Field>
                        <Field name="Usage rights">
                          <input
                            name="license"
                            required
                            defaultValue="Institution-owned"
                          />
                        </Field>
                        <button disabled={busy || !topic}>
                          {uploadingPaper
                            ? "Uploading paper…"
                            : "Upload for preview"}
                        </button>
                      </form>
                      <Field name="Saved imports">
                        <select
                          value={preview?.id ?? ""}
                          onChange={(e) =>
                            setPreview(
                              imports.find((i) => i.id === e.target.value) ??
                                null,
                            )
                          }
                        >
                          <option value="">Select saved preview</option>
                          {imports.map((i) => (
                            <option key={i.id} value={i.id}>
                              {i.created_at.slice(0, 16)} · {i.state} ·{" "}
                              {i.rows.length} rows
                            </option>
                          ))}
                        </select>
                      </Field>
                    </section>
                  )}
                </div>
              </details>
              {tab === "import" && preview && !importActive && (
                <section className="card" ref={importReviewRef}>
                  <h2>Import review · {preview.state}</h2>
                  <button
                    disabled={busy}
                    onClick={() =>
                      void run(
                        async () =>
                          setPreview(await api.detail("imports", preview.id)),
                        "",
                      )
                    }
                  >
                    Reload import preview
                  </button>
                  {preview.asset_id && (
                    <a
                      href={api.url("assets/" + preview.asset_id)}
                      target="_blank"
                      rel="noreferrer"
                    >
                      Open original private source
                    </a>
                  )}
                  {preview.state === "extracting" && (
                    <p>
                      Reading the document. Questions will appear here
                      automatically. Background tasks has errors and retry
                      controls.
                    </p>
                  )}
                  {preview.raw_text && (
                    <DocumentMapping
                      key={preview.id + preview.version}
                      preview={preview}
                      api={api}
                      busy={busy}
                      onParsed={setPreview}
                      run={run}
                    />
                  )}
                  <p>
                    {preview.rows.length} questions detected. Compare the
                    numbering and all options with every source page. Missing
                    keys must be supplied by faculty before commitment.
                  </p>
                  {preview.state === "preview" && preview.rows.length === 0 && (
                    <p role="alert" className="error">
                      No question rows were recognized. Your original source and
                      OCR text are retained. Review and map the transcription,
                      or upload a corrected structured file.
                    </p>
                  )}
                  {preview.state === "preview" && preview.rows.length > 0 && (
                    <>
                      <KeyImportPanel
                        key={preview.id}
                        preview={preview}
                        api={api}
                        jobs={jobs}
                        onApplied={setPreview}
                      />
                      <MarkingRules
                        numbers={preview.rows
                          .filter((r) => !r.committed_id)
                          .map((r) => r.source_number || r.index)}
                        disabled={busy}
                        onApply={async (numbers, marks, penalty) => {
                          await run(async () => {
                            let current = preview;
                            for (const row of preview.rows.filter(
                              (r) =>
                                numbers.includes(r.source_number || r.index) &&
                                !r.committed_id,
                            )) {
                              current = await api.importChange(
                                current.id,
                                "edit",
                                {
                                  index: row.index,
                                  version: current.version,
                                  question: { ...row.question, marks, penalty },
                                },
                              );
                              setPreview(current);
                            }
                          }, "Marking rules saved. Review the questions before saving to the bank.");
                        }}
                      />
                    </>
                  )}
                  {preview.rows.map((row) =>
                    ["single", "multiple"].includes(row.question.type) ? (
                      <ImportChoiceRow
                        key={row.index + "-" + JSON.stringify(row.question)}
                        row={row}
                        onAdvanced={() => {
                          setEdit(row.question);
                          setImportIndex(row.index);
                        }}
                        disabled={busy || preview.state !== "preview"}
                        onSave={(q) =>
                          void run(
                            async () =>
                              setPreview(
                                await api.importChange(preview.id, "edit", {
                                  index: row.index,
                                  version: preview.version,
                                  question: q,
                                }),
                              ),
                            "Reviewed question saved.",
                          )
                        }
                      />
                    ) : (
                      <article className="academic-row" key={row.index}>
                        <span>
                          {row.index}. {row.question.stem}
                        </span>
                        <span className={row.error ? "error" : "muted"}>
                          {row.error ||
                            (row.committed_id
                              ? "Committed"
                              : row.existing_id
                                ? "Already in bank"
                                : "Ready for draft")}
                        </span>
                        <button
                          onClick={() => {
                            setEdit(row.question);
                            setImportIndex(row.index);
                          }}
                          disabled={
                            preview.state !== "preview" || !!row.committed_id
                          }
                        >
                          Correct row {row.index}
                        </button>
                      </article>
                    ),
                  )}
                  <button
                    disabled={
                      busy ||
                      preview.state !== "preview" ||
                      !preview.rows.some((r) => !r.error && !r.committed_id)
                    }
                    onClick={() =>
                      void run(
                        async () =>
                          setPreview(
                            await api.importChange(preview.id, "commit", {
                              version: preview.version,
                              selected: preview.rows
                                .filter((r) => !r.error)
                                .map((r) => r.index),
                            }),
                          ),
                        "Clean rows committed as drafts. Approve them after review.",
                      )
                    }
                  >
                    Save reviewed questions to bank
                  </button>
                </section>
              )}
              {tab === "author" && !edit && (
                <section className="card manual-start">
                  <h2>Write your next question</h2>
                  <p>
                    Choose a topic above, then add the wording, options, correct
                    answer and marks.
                  </p>
                  <button
                    className="primary"
                    disabled={!topic}
                    onClick={() => {
                      setEdit(emptyQuestion(topic));
                      setImportIndex(null);
                    }}
                  >
                    Add a question
                  </button>
                </section>
              )}
              {edit && (
                <section className="card" ref={editorRef}>
                  <h2>
                    {importIndex !== null
                      ? `Correct import row ${importIndex}`
                      : edit.id
                        ? "Create a new question revision"
                        : "Author a question"}
                  </h2>
                  <QuestionEditor
                    key={`${edit.id ?? "new"}-${importIndex}`}
                    initial={edit}
                    api={api}
                    course={chosen?.course_id ?? ""}
                    campus={chosen?.campus_id ?? ""}
                    busy={busy}
                    onCancel={() => {
                      setEdit(null);
                      setImportIndex(null);
                    }}
                    onSave={(q) =>
                      void run(async () => {
                        if (importIndex !== null && preview)
                          setPreview(
                            await api.importChange(preview.id, "edit", {
                              index: importIndex,
                              version: preview.version,
                              question: q,
                            }),
                          );
                        else
                          await api.question(
                            { question: q, version: edit.version },
                            edit.id,
                          );
                        setEdit(null);
                        setImportIndex(null);
                      }, "Question saved as a draft. Published papers retain their original revision.")
                    }
                  />
                </section>
              )}
              {tab === "curriculum" && (
                <section className="card">
                  <div className="section-heading">
                    <h2>Question bank {chosen ? `/ ${chosen.name}` : ""}</h2>
                    <button
                      disabled={!topic}
                      onClick={() => {
                        setEdit(emptyQuestion(topic));
                        setImportIndex(null);
                      }}
                    >
                      Author question
                    </button>
                  </div>
                  {bank.length > 0 && (
                    <MarkingRules
                      numbers={bank.map((_, i) => i + 1)}
                      disabled={busy}
                      onApply={async (numbers, marks, penalty) => {
                        await run(async () => {
                          for (const number of numbers) {
                            const q = bank[number - 1];
                            const updated = await api.question(
                              {
                                question: { ...q, marks, penalty },
                                version: q.version,
                              },
                              q.id,
                            );
                            const saved =
                              "question" in updated
                                ? updated.question
                                : updated;
                            setBank((current) =>
                              current.map((old) =>
                                old.id === q.id ? saved : old,
                              ),
                            );
                          }
                        }, "Marks saved as new drafts. Review and approve the changed questions before use.");
                      }}
                    />
                  )}
                  <div className="bank-tools">
                    <Field name="Search questions">
                      <input
                        type="search"
                        value={bankQuery}
                        onChange={(e) => setBankQuery(e.target.value)}
                        placeholder="Search question wording"
                      />
                    </Field>
                    <Field name="Question status">
                      <select
                        value={bankFilter}
                        onChange={(e) => setBankFilter(e.target.value)}
                      >
                        <option value="all">All questions</option>
                        <option value="draft">Needs review</option>
                        <option value="approved">Approved for tests</option>
                      </select>
                    </Field>
                    <button onClick={() => setTab("generate")}>
                      Create with AI
                    </button>
                    <button onClick={() => setTab("import")}>
                      Upload questions
                    </button>
                  </div>
                  {bank.length === 0 && (
                    <p className="muted">No questions in this topic yet.</p>
                  )}
                  {bank
                    .filter(
                      (q) =>
                        (bankFilter === "all" || q.state === bankFilter) &&
                        q.stem.toLowerCase().includes(bankQuery.toLowerCase()),
                    )
                    .map((q) => (
                      <article className="question-card" key={q.id}>
                        <div className="section-heading">
                          <strong>
                            Question {bank.indexOf(q) + 1} · {q.reference} ·{" "}
                            {q.state}
                          </strong>
                          <span>
                            {q.type} · +{q.marks} / −{q.penalty}
                          </span>
                        </div>
                        <details className="question-preview">
                          <summary>{q.stem}</summary>
                          <QuestionContent q={q} api={api} showStem={false} />
                          <details>
                            <summary>Key, solution and provenance</summary>
                            <p>Key: {q.key?.join(", ") || answerRuleText(q)}</p>
                            <p>{q.solution}</p>
                            <p className="muted">
                              {q.origin} · {q.provenance} · {q.license}
                            </p>
                          </details>
                        </details>
                        <div className="actions">
                          <button
                            onClick={() => {
                              setEdit(q);
                              setImportIndex(null);
                            }}
                          >
                            Revise {q.reference}
                          </button>
                          {q.state !== "approved" && (
                            <button
                              disabled={busy}
                              onClick={() =>
                                void run(
                                  () =>
                                    api.review(q.id!, {
                                      version: q.version!,
                                      state: "approved",
                                      note: "Teacher checked wording, answer key, solution and rights",
                                    }),
                                  "Approved for assessment composition.",
                                )
                              }
                            >
                              Approve {q.reference}
                            </button>
                          )}
                          {q.state === "approved" && (
                            <label className="check">
                              <input
                                type="checkbox"
                                checked={selected.includes(q.id!)}
                                onChange={(e) =>
                                  setSelected(
                                    e.target.checked
                                      ? [...selected, q.id!]
                                      : selected.filter((id) => id !== q.id),
                                  )
                                }
                              />
                              Select for paper
                            </label>
                          )}
                        </div>
                      </article>
                    ))}
                  {selected.length > 0 && (
                    <button onClick={() => setTab("papers")}>
                      Compose with {selected.length} selected questions
                    </button>
                  )}
                </section>
              )}
              {tab === "import" &&
                preview?.rows.some((r) => r.committed_id) && (
                  <button
                    className="primary"
                    onClick={() => setTab("curriculum")}
                  >
                    Next: review imported questions
                  </button>
                )}
            </>
          )}
          {tab === "papers" && (
            <>
              <div className="academic-columns">
                <section className="card">
                  {selected.length === 0 && (
                    <div className="empty-message">
                      <p>
                        Start by selecting the questions you want to include.
                      </p>
                      <button onClick={() => setTab("curriculum")}>
                        Choose questions from the bank
                      </button>
                    </div>
                  )}
                  <h2>Create a new assessment</h2>
                  <p>
                    Select approved questions in Topics & questions. Their order
                    below is the printed and online order.
                  </p>
                  <ol>
                    {selected.map((id, i) => (
                      <li key={id}>
                        {bank.find((q) => q.id === id)?.stem ?? id}
                        <button
                          type="button"
                          disabled={i === 0}
                          onClick={() =>
                            setSelected((s) => {
                              const next = [...s];
                              [next[i - 1], next[i]] = [next[i], next[i - 1]];
                              return next;
                            })
                          }
                        >
                          Move up
                        </button>
                        <button
                          onClick={() =>
                            setSelected((s) => s.filter((x) => x !== id))
                          }
                        >
                          Remove
                        </button>
                      </li>
                    ))}
                  </ol>
                  <form
                    onSubmit={(e) => {
                      const f = values(e);
                      void run(async () => {
                        if (!chosen) throw new Error("Select a topic first.");
                        const policy = str(f, "policy");
                        const p = await api.compose({
                          campus_id: chosen.campus_id,
                          course_id: chosen.course_id,
                          title: str(f, "title"),
                          purpose: str(f, "purpose"),
                          instructions: str(f, "instructions"),
                          duration_minutes: Number(str(f, "duration")),
                          release_policy: str(f, "release"),
                          sections: [
                            {
                              id: "core",
                              title: "Core questions",
                              policy,
                              count:
                                policy === "all"
                                  ? selected.length
                                  : Number(str(f, "count")),
                              cap: str(f, "cap"),
                              question_ids: selected,
                            },
                          ],
                        });
                        setPaper(p);
                      }, "Draft paper saved with pinned question revisions.");
                    }}
                  >
                    <Field name="Assessment title">
                      <input name="title" required minLength={3} />
                    </Field>
                    <Field name="Purpose">
                      <select name="purpose">
                        {[
                          "practice",
                          "quiz",
                          "baseline",
                          "mock",
                          "formal",
                          "homework",
                        ].map((p) => (
                          <option key={p}>{p}</option>
                        ))}
                      </select>
                    </Field>
                    <Field name="Instructions">
                      <textarea
                        name="instructions"
                        defaultValue="Read each question. Select or enter your response. Blank answers receive no penalty."
                      />
                    </Field>
                    <Field name="Duration in minutes">
                      <input
                        name="duration"
                        type="number"
                        min={1}
                        max={480}
                        defaultValue={30}
                      />
                    </Field>
                    <Field name="Result release">
                      <select name="release">
                        <option value="staff">
                          Staff release after the assignment window
                        </option>
                        <option value="immediate">
                          Immediate, formative work only
                        </option>
                      </select>
                    </Field>
                    <details>
                      <summary>
                        Advanced scoring: optional questions and section cap
                      </summary>{" "}
                      <Field name="Section contribution">
                        <select name="policy">
                          <option value="all">All questions</option>
                          <option value="first">First N attempted</option>
                          <option value="best">Best N attempted</option>
                        </select>
                      </Field>
                      <Field name="Required contribution count for optional section">
                        <input
                          type="number"
                          name="count"
                          min={1}
                          defaultValue={1}
                        />
                      </Field>
                      <Field name="Optional section mark cap">
                        <input name="cap" inputMode="decimal" />
                      </Field>
                    </details>{" "}
                    <button disabled={busy || selected.length === 0}>
                      Save paper draft
                    </button>
                  </form>
                </section>
                <section className="card">
                  <h2>Your assessments</h2>
                  {papers.map((p) => (
                    <button
                      className="paper-choice"
                      key={p.id}
                      onClick={() => setPaper(p)}
                    >
                      {p.reference} / {p.spec.title}
                      <small>
                        {p.state} · {p.maximum} marks
                      </small>
                    </button>
                  ))}
                </section>
              </div>
              {paper && (
                <PaperOperations
                  key={paper.id + paper.version}
                  paper={paper}
                  api={api}
                  busy={busy}
                  run={run}
                  onUpdated={setPaper}
                  onSitting={setSitting}
                />
              )}
            </>
          )}
          {tab === "scan" && (
            <>
              <OMRCheckerReader api={api} courses={courses} />
              <details className="card">
                <summary>Advanced pixel-reader tools</summary>
                <GridInspection api={api} courses={courses} />
              </details>
              <Field name="Published assessment">
                <select
                  value={paper?.id ?? ""}
                  onChange={(e) =>
                    setPaper(
                      papers.find((p) => p.id === e.target.value) ?? null,
                    )
                  }
                >
                  <option value="">Select paper</option>
                  {(paper && !papers.some((p) => p.id === paper.id)
                    ? [paper, ...papers]
                    : papers
                  )
                    .filter((p) => p.state === "published")
                    .map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.reference} / {p.spec.title}
                      </option>
                    ))}
                </select>
              </Field>
              {paper?.state === "published" && (
                <ScanDesk
                  key={paper.id}
                  paper={paper}
                  api={api}
                  run={run}
                  busy={busy}
                  onSitting={setSitting}
                />
              )}
            </>
          )}
          {tab === "jobs" && (
            <section className="card">
              <div className="section-heading">
                <h2>Processing & recovery</h2>
                <button disabled={busy} onClick={() => void run(refresh, "")}>
                  Refresh tasks
                </button>
              </div>
              <p>
                Jobs retain failed input and partial evidence. Correct
                configuration or input before retrying. Provider output is
                labelled by its actual source.
              </p>
              {jobs.length === 0 && <p>No academic jobs yet.</p>}
              {jobs.map((j) => (
                <article className="question-card" key={j.id}>
                  <strong>
                    {(
                      {
                        generate_questions: "Question generation",
                        recognize_scan: "Answer-sheet reading",
                        extract_document: "Paper extraction",
                        tutor: "Tutor reply",
                      } as Record<string, string>
                    )[j.kind] ?? j.kind}{" "}
                    · {j.state.replaceAll("_", " ")}
                  </strong>
                  <progress
                    max={100}
                    value={j.progress}
                    aria-label="Job progress"
                  />
                  <p>
                    {j.attempts} attempts ·{" "}
                    {j.provider || "Provider not recorded"} {j.model}
                  </p>
                  {j.provider === "development_fixture" && (
                    <p className="environment">
                      Deterministic demonstration fixture. No external AI
                      generation or topic-quality validation.
                    </p>
                  )}
                  {j.error && <p className="error">{j.error}</p>}
                  <div className="actions">
                    {["failed", "configuration_required", "cancelled"].includes(
                      j.state,
                    ) && (
                      <button
                        disabled={busy}
                        onClick={() =>
                          void run(
                            () => api.jobAction(j.id, "retry"),
                            "Job queued for retry.",
                          )
                        }
                      >
                        Retry job
                      </button>
                    )}
                    {[
                      "queued",
                      "running",
                      "failed",
                      "configuration_required",
                    ].includes(j.state) && (
                      <button
                        disabled={busy}
                        onClick={() =>
                          void run(
                            () => api.jobAction(j.id, "cancel"),
                            "Job cancelled; existing evidence retained.",
                          )
                        }
                      >
                        Cancel job
                      </button>
                    )}
                  </div>
                </article>
              ))}
            </section>
          )}
        </>
      ) : (
        <section className="card">
          <h2>Your assessments</h2>
          <p>
            Deadlines use server time. Resume transfers the active writer to
            this signed-in session.
          </p>
          {assignments.length === 0 && <p>No assigned assessments.</p>}
          {assignments.map((a) => (
            <article className="academic-row" key={a.id}>
              <div>
                <strong>{a.title}</strong>
                <p>
                  {a.reference} · {a.state.replaceAll("_", " ")}
                </p>
                <small>
                  {new Date(a.opens_at).toLocaleString()} →{" "}
                  {new Date(a.closes_at).toLocaleString()}
                </small>
              </div>
              <button
                disabled={busy}
                onClick={() =>
                  void run(
                    async () => setSitting(await api.start(a.id, true)),
                    "",
                  )
                }
              >
                {a.state === "not_started"
                  ? "Start assessment"
                  : a.state === "active"
                    ? "Resume here"
                    : "View outcome"}
              </button>
            </article>
          ))}
        </section>
      )}
    </section>
  );
}
function DocumentMapping({
  preview,
  api,
  busy,
  onParsed,
  run,
}: {
  preview: A<"AcademicImport">;
  api: EducationClient;
  busy: boolean;
  onParsed: (p: A<"AcademicImport">) => void;
  run: Run;
}) {
  const [text, setText] = useState(preview.raw_text);
  return (
    <details open={preview.rows.length === 0 || undefined}>
      <summary>Review OCR text and map questions</summary>
      <p>
        Compare every question and key with the source. Supported text format:
        numbered questions, A. option lines, Answer: A (or A|C), and Solution:
        explanation. Missing keys remain invalid. The original OCR text is
        retained.
      </p>
      <form
        onSubmit={(e) => {
          const f = values(e);
          void run(
            async () =>
              onParsed(
                await api.parseDocument(preview.id, {
                  text,
                  provenance: str(f, "source"),
                  license: str(f, "license"),
                  version: preview.version,
                }),
              ),
            "Mapped transcription saved for row review.",
          );
        }}
      >
        <Field name="Reviewed transcription">
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            rows={15}
          />
        </Field>
        <Field name="Document provenance">
          <input
            name="source"
            required
            defaultValue="Teacher-reviewed document transcription"
          />
        </Field>
        <Field name="Document usage rights">
          <input name="license" required defaultValue="Institution-owned" />
        </Field>
        <button disabled={busy || preview.state !== "preview"}>
          Map reviewed text to draft rows
        </button>
      </form>
    </details>
  );
}
function answerRuleText(q: Q): string {
  if (q.key?.length) return q.key.join(" + ");
  if (q.numeric) {
    const rule = q.numeric;
    const parts: string[] = [];
    if (rule.values?.length)
      parts.push("Accepted: " + rule.values.join(" or "));
    if (rule.min !== undefined && rule.min !== "")
      parts.push(`Range: ${rule.min} to ${rule.max}`);
    if (rule.absolute) parts.push("allowed difference " + rule.absolute);
    if (rule.relative) parts.push("relative tolerance " + rule.relative);
    if (rule.integer) parts.push("whole numbers only");
    return parts.join("; ");
  }
  return (q.rubric ?? [])
    .map((c) => `${c.description} (${c.marks} marks)`)
    .join("; ");
}
function QuestionContent({
  q,
  api,
  options = true,
  showStem = true,
}: {
  showStem?: boolean;
  q: Q;
  api: EducationClient;
  options?: boolean;
}) {
  return (
    <>
      {showStem && <p className="question-stem">{q.stem}</p>}
      {q.asset_id && (
        <img
          className="question-image"
          src={api.url("assets/" + q.asset_id)}
          alt={q.alt ?? ""}
        />
      )}
      {options && (
        <ol type="A">
          {q.options?.map((o) => (
            <li key={o.id}>
              {o.text}
              {o.asset_id && (
                <img
                  className="question-image"
                  src={api.url("assets/" + o.asset_id)}
                  alt={o.alt ?? ""}
                />
              )}
            </li>
          ))}
        </ol>
      )}
    </>
  );
}
type Run = (fn: () => Promise<unknown>, message?: string) => Promise<void>;
function PaperOperations({
  paper,
  api,
  busy,
  run,
  onUpdated,
  onSitting,
}: {
  paper: Paper;
  api: EducationClient;
  busy: boolean;
  run: Run;
  onUpdated: (p: Paper) => void;
  onSitting: (s: Sitting) => void;
}) {
  const [recipients, setRecipients] = useState<A<"AcademicRecipient">[]>([]);
  const [assigned, setAssigned] = useState<A<"AcademicAssigned">[]>([]);
  const [chosen, setChosen] = useState<string[]>([]);
  useEffect(() => {
    let active = true;
    Promise.all([
      api.recipients(paper.id),
      api.list("assignments", { paper: paper.id }),
    ])
      .then(([r, a]) => {
        if (active) {
          setRecipients(r);
          setAssigned(a);
        }
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, [api, paper.id]);
  return (
    <section className="card">
      <div className="section-heading">
        <h2>
          {paper.reference} / {paper.spec.title}
        </h2>
        <span>
          {paper.state} · {paper.maximum} marks
        </span>
      </div>
      <p>{paper.spec.instructions}</p>
      <details>
        <summary>Review pinned paper and key</summary>
        {paper.spec.sections.map((s) => (
          <section key={s.id}>
            <h3>
              {s.title} · {s.policy} / {s.count}
            </h3>
            {s.questions.map((q) => (
              <article className="question-card" key={q.id}>
                <QuestionContent q={q} api={api} />
                <p>Key: {q.key?.join(", ") || answerRuleText(q)}</p>
                <p>{q.solution}</p>
              </article>
            ))}
          </section>
        ))}
      </details>
      {paper.state === "draft" ? (
        <button
          disabled={busy}
          onClick={() =>
            void run(async () => {
              await api.publish(paper.id, { version: paper.version });
              onUpdated(await api.detail("papers", paper.id));
            }, "Published. Content and scoring are now immutable.")
          }
        >
          Publish reviewed paper
        </button>
      ) : (
        <form
          onSubmit={(e) => {
            const f = values(e);
            void run(async () => {
              await api.assign({
                paper_id: paper.id,
                learners: chosen,
                opens_at: new Date(str(f, "opens")).toISOString(),
                closes_at: new Date(str(f, "closes")).toISOString(),
                extra_minutes: Number(str(f, "extra")),
              });
              setAssigned(await api.list("assignments", { paper: paper.id }));
            }, "Assigned to the selected learners.");
          }}
        >
          <h3>Assign and accommodate</h3>
          <fieldset>
            <legend>Eligible learners</legend>
            {recipients.length === 0 && (
              <p>No active course enrollments. Add learners in Admissions.</p>
            )}
            {recipients.map((r) => (
              <label className="check" key={r.id}>
                <input
                  type="checkbox"
                  checked={chosen.includes(r.id)}
                  onChange={(e) =>
                    setChosen(
                      e.target.checked
                        ? [...chosen, r.id]
                        : chosen.filter((id) => id !== r.id),
                    )
                  }
                />
                {r.name} / {r.reference}
              </label>
            ))}
          </fieldset>
          <Field name="Opens (your device timezone)">
            <input
              type="datetime-local"
              name="opens"
              defaultValue={nowLocal()}
              required
            />
          </Field>
          <Field name="Closes (your device timezone)">
            <input type="datetime-local" name="closes" required />
          </Field>
          <Field name="Approved extra minutes">
            <input
              name="extra"
              type="number"
              min={0}
              max={480}
              defaultValue={0}
            />
          </Field>
          <button disabled={busy || chosen.length === 0}>Assign paper</button>
        </form>
      )}
      {paper.state === "published" && (
        <div className="actions">
          <a
            className="button-link"
            href={api.url("print-paper/" + paper.id)}
            target="_blank"
            rel="noreferrer"
          >
            Open printable question booklet
          </a>
          <a
            className="button-link"
            href={api.url("print-paper/" + paper.id + "?answers=1")}
            target="_blank"
            rel="noreferrer"
          >
            Open approved staff marking key
          </a>
        </div>
      )}
      <h3>Assigned evidence</h3>
      {assigned.map((a) => (
        <article className="academic-row" key={a.id}>
          <span>
            {a.learner} · {a.state}
          </span>
          {a.sitting_id && (
            <button
              onClick={() =>
                void run(
                  async () =>
                    onSitting(await api.detail("sittings", a.sitting_id)),
                  "",
                )
              }
            >
              Review evidence
            </button>
          )}
        </article>
      ))}
    </section>
  );
}
function ScanDesk({
  paper,
  api,
  run,
  busy,
  onSitting,
}: {
  paper: Paper;
  api: EducationClient;
  run: Run;
  busy: boolean;
  onSitting: (s: Sitting) => void;
}) {
  const [layouts, setLayouts] = useState<A<"AcademicSheetLayout">[]>([]);
  const [layout, setLayout] = useState("");
  const [assignments, setAssignments] = useState<A<"AcademicAssigned">[]>([]);
  const [assignment, setAssignment] = useState("");
  const [batches, setBatches] = useState<A<"AcademicBatch">[]>([]);
  const [batch, setBatch] = useState("");
  const [sheets, setSheets] = useState<A<"AcademicScanSheet">[]>([]);
  const [job, setJob] = useState<A<"AcademicJob"> | null>(null);
  const [error, setError] = useState("");
  const [loadingSheets, setLoadingSheets] = useState(true);
  const [uploadingScan, setUploadingScan] = useState(false);
  const [uploadMode, setUploadMode] = useState("individual");
  const [scanJobs, setScanJobs] = useState<A<"AcademicJob">[]>([]);
  const [uploadFailures, setUploadFailures] = useState<string[]>([]);
  const [scanAt, setScanAt] = useState(nowLocal());
  const refresh = useCallback(async () => {
    const [l, a, b] = await Promise.all([
      api.list("layouts", { paper: paper.id }),
      api.list("assignments", { paper: paper.id }),
      api.list("batches", { paper: paper.id }),
    ]);
    setLayouts(l);
    setLayout((current) =>
      l.some((v) => v.id === current) ? current : (l[0]?.id ?? ""),
    );
    setAssignments(a);
    setAssignment((current) =>
      a.some((v) => v.id === current) ? current : a.length === 1 ? a[0].id : "",
    );
    setBatches(b);
    setLoadingSheets(false);
  }, [api, paper.id]);
  useEffect(() => {
    refresh().catch((e) => setError(e.message));
  }, [refresh]);
  const load = useCallback(async () => {
    if (batch) setSheets(await api.list("scan-sheets", { batch }));
    else setSheets([]);
  }, [api, batch]);
  useEffect(() => {
    let active = true;
    setScanJobs([]);
    setJob(null);
    if (batch)
      void api
        .list("jobs")
        .then((jobs) => {
          if (!active) return;
          const matching = jobs.filter(
            (j) =>
              j.kind === "recognize_scan" &&
              (j.input as Record<string, unknown>)?.batch_id === batch,
          );
          setScanJobs((current) => [
            ...current,
            ...matching.filter(
              (j) => !current.some((saved) => saved.id === j.id),
            ),
          ]);
          setJob((current) => current ?? matching[0] ?? null);
        })
        .catch((e) => {
          if (active) setError(e.message);
        });
    return () => {
      active = false;
    };
  }, [api, batch]);
  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, [load]);
  useEffect(() => {
    const activeJobs = scanJobs.filter((j) =>
      ["queued", "running"].includes(j.state),
    );
    if (!activeJobs.length) return;
    let active = true,
      pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        const results = await Promise.allSettled(
          activeJobs.map((j) => api.detail("jobs", j.id)),
        );
        if (!active) return;
        const updated = results.flatMap((r) =>
          r.status === "fulfilled" ? [r.value] : [],
        );
        setScanJobs((current) =>
          current.map((j) => updated.find((next) => next.id === j.id) ?? j),
        );
        setJob(
          (current) => updated.find((j) => j.id === current?.id) ?? current,
        );
        if (updated.some((j) => j.state === "completed")) await load();
        if (results.some((r) => r.status === "rejected"))
          setError(
            "Connection interrupted. Saved uploads are retained; status will retry automatically.",
          );
      } catch (e) {
        if (active)
          setError(
            e instanceof Error
              ? e.message
              : "Could not refresh scan review. Retry from the review panel.",
          );
      } finally {
        pending = false;
      }
    }, 2500);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, scanJobs, load]);
  return (
    <>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <ol className="scan-steps" aria-label="Scan workflow">
        <li>Get matching sheet</li>
        <li>Upload completed sheet</li>
        <li>Check and save result</li>
      </ol>
      <div className="academic-columns">
        <section className="card">
          <h2>1. Get a matching answer sheet</h2>
          {loadingSheets && (
            <p role="status">Loading the assigned learners and sheet format…</p>
          )}
          <p>
            The sheet must match this assessment and its assigned learner. Keep
            the black corner squares visible when scanning.
          </p>
          {!loadingSheets && assignments.length === 0 && (
            <p className="notice">
              Assign this test to a learner from Assessments before printing or
              uploading.
            </p>
          )}
          <Field name="Assigned learner">
            <select
              value={assignment}
              onChange={(e) => setAssignment(e.target.value)}
            >
              <option value="">Choose learner</option>
              {assignments.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.learner} / {a.state.replaceAll("_", " ")}
                </option>
              ))}
            </select>
          </Field>
          {layouts.length > 0 && (
            <Field name="Sheet format">
              <select
                value={layout}
                onChange={(e) => {
                  setLayout(e.target.value);
                  setBatch("");
                }}
              >
                {layouts.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.name}
                  </option>
                ))}
              </select>
            </Field>
          )}
          {layout && assignment && (
            <a
              className="button-link"
              target="_blank"
              rel="noreferrer"
              href={api.url(`sheet-image/${layout}?assignment=${assignment}`)}
            >
              Open printable answer sheet
            </a>
          )}
          {paper.spec.title.startsWith("Sample:") && assignment && (
            <a
              href={api.url(
                `sample-files/${paper.id}?assignment=${assignment}`,
              )}
              download
            >
              Download matching filled sample sheets (.zip)
            </a>
          )}
          <details open={layouts.length === 0 || undefined}>
            <summary>
              {layouts.length
                ? "Change or create another sheet format"
                : "Prepare the sheet format"}
            </summary>
            <form
              onSubmit={(e) => {
                const f = values(e);
                void run(async () => {
                  const l = await api.layout({
                    paper_id: paper.id,
                    columns: Number(str(f, "columns")),
                  });
                  await refresh();
                  setLayout(l.id);
                  setBatch("");
                }, "Sheet format ready. Choose the learner, then open the printable sheet.");
              }}
            >
              <Field name="Sheet arrangement">
                <select name="columns">
                  <option value="1">One column / up to 40 questions</option>
                  <option value="2">Two columns / up to 80 questions</option>
                </select>
              </Field>
              <button disabled={busy}>Create sheet layout</button>
            </form>
          </details>
        </section>
        <section className="card">
          <h2>2. Upload completed OMR sheets</h2>
          <fieldset className="upload-mode">
            <legend>Upload mode</legend>
            <label className="check">
              <input
                type="radio"
                name="upload-mode"
                checked={uploadMode === "individual"}
                disabled={uploadingScan}
                onChange={() => setUploadMode("individual")}
              />
              Individual sheet
            </label>
            <label className="check">
              <input
                type="radio"
                name="upload-mode"
                checked={uploadMode === "batch"}
                disabled={uploadingScan}
                onChange={() => setUploadMode("batch")}
              />
              Batch upload
            </label>
          </fieldset>
          <p className="muted">
            Registered sheet IDs match each page to its assigned student.
            Unreadable identities stay in review.
          </p>
          <p>
            Use a scanned PDF or a clear PNG/JPEG photograph. We will read the
            marks and ask you to check anything uncertain.
          </p>
          <form
            onSubmit={(e) => {
              const f = values(e);
              void run(async () => {
                setUploadingScan(true);
                setUploadFailures([]);
                try {
                  let id = batch;
                  const existing = batches.find((b) => b.id === id);
                  if (!existing || existing.layout_id !== layout) {
                    const b = await api.batch({
                      paper_id: paper.id,
                      layout_id: layout,
                      sitting_at: new Date(scanAt).toISOString(),
                    });
                    id = b.id;
                    setBatch(id);
                    await refresh();
                  }
                  const files = f
                    .getAll("file")
                    .filter(
                      (file): file is File =>
                        file instanceof File && file.size > 0,
                    );
                  if (files.length > 50)
                    throw new Error("Choose up to 50 files per batch.");
                  for (const file of files) {
                    const upload = new FormData();
                    upload.set("file", file);
                    try {
                      const next = await api.scanUpload(id, upload);
                      setJob(next);
                      setScanJobs((current) =>
                        current.some((j) => j.id === next.id)
                          ? current.map((j) => (j.id === next.id ? next : j))
                          : [...current, next],
                      );
                    } catch (e) {
                      setUploadFailures((current) => [
                        ...current,
                        `${file.name}: ${e instanceof Error ? e.message : "Upload failed"}`,
                      ]);
                    }
                  }
                } finally {
                  setUploadingScan(false);
                }
              }, "");
            }}
          >
            <Field name="Actual test date and time">
              <input
                type="datetime-local"
                value={scanAt}
                required
                onChange={(e) => {
                  setScanAt(e.target.value);
                  setBatch("");
                }}
              />
            </Field>
            <Field name="Scanned PDF or page image">
              <input
                key={uploadMode}
                name="file"
                type="file"
                multiple={uploadMode === "batch"}
                accept="application/pdf,image/png,image/jpeg"
                required
              />
            </Field>
            <small className="muted">
              {uploadMode === "batch"
                ? "Choose up to 50 files. "
                : "Choose one image or PDF. "}
              Up to 50 MiB per file and 50 pages per PDF.
            </small>
            <button
              className="primary"
              disabled={
                !layout ||
                uploadingScan ||
                busy ||
                (!!job && ["queued", "running"].includes(job.state))
              }
            >
              {uploadingScan
                ? "Uploading sheet…"
                : job && ["queued", "running"].includes(job.state)
                  ? "Sheet reading in progress…"
                  : uploadMode === "batch"
                    ? "Upload and read batch"
                    : "Upload and read sheet"}
            </button>
          </form>
          {uploadFailures.length > 0 && (
            <div className="error" role="alert">
              <p>
                Some files were not uploaded. Select these files again to retry;
                saved uploads are retained.
              </p>
              <ul>
                {uploadFailures.map((failure, i) => (
                  <li key={i}>{failure}</li>
                ))}
              </ul>
            </div>
          )}
          {scanJobs.length > 1 && (
            <div className="batch-status" aria-live="polite">
              <h3>Batch progress</h3>
              {scanJobs.map((j, i) => (
                <div key={j.id}>
                  <span>Upload {i + 1}</span>
                  <span>{j.state.replaceAll("_", " ")}</span>
                  {["failed", "configuration_required", "cancelled"].includes(
                    j.state,
                  ) && (
                    <button
                      disabled={busy}
                      onClick={() =>
                        void run(async () => {
                          await api.jobAction(j.id, "retry");
                          const next = await api.detail("jobs", j.id);
                          setScanJobs((current) =>
                            current.map((v) => (v.id === j.id ? next : v)),
                          );
                        }, "Upload queued again.")
                      }
                    >
                      Retry upload {i + 1}
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
          {(uploadingScan || job) && (
            <TaskProgress
              state={uploadingScan ? "uploading" : job!.state}
              progress={uploadingScan ? undefined : job?.progress}
              title={
                uploadingScan
                  ? "Saving your answer sheets"
                  : job!.state === "completed"
                    ? "Sheet reading finished — review below"
                    : job!.state === "queued"
                      ? "Upload saved — waiting to read"
                      : job!.state === "running"
                        ? "Reading your uploaded sheet…"
                        : "Sheet reading needs attention"
              }
              detail={
                uploadingScan
                  ? "Keep this tab open until the upload is saved."
                  : job!.error ||
                    (job!.state === "completed"
                      ? "Confirm the learner and any uncertain marks below before saving results."
                      : "Your file is saved. Responses are being detected; no result is committed until you review the evidence.")
              }
            >
              {job &&
                ["failed", "configuration_required", "cancelled"].includes(
                  job.state,
                ) && (
                  <button
                    disabled={busy}
                    onClick={() =>
                      void run(async () => {
                        await api.jobAction(job!.id, "retry");
                        const next = await api.detail("jobs", job!.id);
                        setJob(next);
                        setScanJobs((current) =>
                          current.map((j) => (j.id === next.id ? next : j)),
                        );
                      }, "Reading queued again. Your original upload is retained.")
                    }
                  >
                    Retry reading this upload
                  </button>
                )}
            </TaskProgress>
          )}
          <details>
            <summary>Resume an earlier upload group</summary>
            <Field name="Earlier upload group">
              <select
                value={batch}
                onChange={(e) => {
                  setBatch(e.target.value);
                  setJob(null);
                  setScanJobs([]);
                  const b = batches.find((b) => b.id === e.target.value);
                  if (b) {
                    setLayout(b.layout_id);
                    const d = new Date(b.sitting_at);
                    setScanAt(
                      new Date(d.getTime() - d.getTimezoneOffset() * 60000)
                        .toISOString()
                        .slice(0, 16),
                    );
                  }
                }}
              >
                <option value="">Start a new upload group</option>
                {batches.map((b) => (
                  <option key={b.id} value={b.id}>
                    {new Date(b.sitting_at).toLocaleString()} /{" "}
                    {b.id.slice(0, 8)}
                  </option>
                ))}
              </select>
            </Field>
          </details>
        </section>
      </div>
      <section className="card">
        <div className="section-heading">
          <h2>3. Check answers and save the result</h2>
          <button disabled={!batch || busy} onClick={() => void run(load, "")}>
            Refresh scan review
          </button>
        </div>
        {sheets.length === 0 && (
          <p className="empty-message">
            {job && ["queued", "running"].includes(job.state)
              ? "Your uploaded sheet is being processed. This review updates automatically."
              : "Upload a completed sheet to see the original page and detected answers here."}
          </p>
        )}
        {sheets.map((sheet) => (
          <ScanReview
            key={sheet.id + sheet.version}
            sheet={sheet}
            paper={paper}
            api={api}
            assignments={assignments}
            busy={busy}
            onSave={(data) =>
              void run(
                async () => {
                  const result = await api.scanReview(sheet.id, data);
                  await load();
                  if (result.sitting_id)
                    onSitting(await api.detail("sittings", result.sitting_id));
                },
                data.commit
                  ? "Result saved and graded using the published answer key."
                  : "Review saved. You can continue later.",
              )
            }
          />
        ))}
      </section>
    </>
  );
}
function ScanReview({
  sheet,
  paper,
  api,
  assignments,
  busy,
  onSave,
}: {
  sheet: A<"AcademicScanSheet">;
  paper: Paper;
  api: EducationClient;
  assignments: A<"AcademicAssigned">[];
  busy: boolean;
  onSave: (data: A<"AcademicScanReview">) => void;
}) {
  const [assignment, setAssignment] = useState(sheet.assignment_id);
  const [confirmed, setConfirmed] = useState(sheet.identity_confirmed);
  const [corrections, setCorrections] = useState(sheet.corrections ?? {});
  const [reason, setReason] = useState("");
  const [commit, setCommit] = useState(false);
  const questions = paper.spec.sections.flatMap((s) => s.questions);
  return (
    <article className="scan-review">
      <h3>
        Page {sheet.page} · {sheet.state}
      </h3>
      <div className="scan-columns">
        <a
          href={api.url("assets/" + sheet.asset_id)}
          target="_blank"
          rel="noreferrer"
        >
          <img
            src={api.url("assets/" + sheet.asset_id)}
            alt={`Original scan page ${sheet.page}; open at full size for review`}
          />
        </a>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            onSave({
              assignment_id: assignment,
              identity_confirmed: confirmed,
              reason,
              version: sheet.version,
              corrections,
              commit,
            });
          }}
        >
          <fieldset disabled={busy || sheet.state === "committed"}>
            <legend>Identity and responses</legend>
            {sheet.machine.error && (
              <p className="error">{sheet.machine.error}</p>
            )}
            <Field name="Confirmed assigned learner">
              <select
                required
                value={assignment}
                onChange={(e) => {
                  setAssignment(e.target.value);
                  setConfirmed(false);
                }}
              >
                <option value="">Select learner</option>
                {assignments.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.learner}
                  </option>
                ))}
              </select>
            </Field>
            <label className="check">
              <input
                type="checkbox"
                checked={confirmed}
                onChange={(e) => setConfirmed(e.target.checked)}
              />
              I checked the sheet identity against the assigned learner
            </label>
            {questions.map((q, i) => {
              const read = sheet.machine.rows?.[q.id!];
              const a = corrections[q.id!] ?? read?.answer ?? blank();
              const unresolved =
                !corrections[q.id!] &&
                (read?.uncertain || !read || !!sheet.machine.error);
              return (
                <details key={q.id} open={unresolved}>
                  <summary>
                    {i + 1}. {q.stem.slice(0, 80)} ·{" "}
                    {unresolved
                      ? "Review required"
                      : corrections[q.id!]
                        ? "Reviewed"
                        : a.selection?.join(", ") ||
                          a.number ||
                          a.text ||
                          "Blank"}
                  </summary>
                  <p className="muted">
                    {read?.reason}{" "}
                    {read?.fill?.length
                      ? `Fill measurements: ${read.fill.map((n) => n.toFixed(2)).join(" / ")}`
                      : ""}
                  </p>
                  <AnswerControl
                    api={api}
                    q={q}
                    answer={a}
                    onChange={(v) =>
                      setCorrections({ ...corrections, [q.id!]: v })
                    }
                  />
                  <label className="check">
                    <input
                      type="checkbox"
                      checked={a.invalid}
                      onChange={(e) =>
                        setCorrections({
                          ...corrections,
                          [q.id!]: { ...a, invalid: e.target.checked },
                        })
                      }
                    />
                    Reviewer confirms an invalid response
                  </label>
                  <button
                    type="button"
                    onClick={() =>
                      setCorrections({ ...corrections, [q.id!]: a })
                    }
                  >
                    Confirm this response, including blank
                  </button>
                </details>
              );
            })}
            <Field name="Review reason / correction explanation">
              <textarea
                required
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            </Field>
            <div className="actions">
              <button onClick={() => setCommit(false)}>Save review</button>
              <button onClick={() => setCommit(true)}>
                Commit reviewed evidence
              </button>
            </div>
          </fieldset>
        </form>
      </div>
    </article>
  );
}
function AnswerControl({
  api,
  q,
  answer,
  onChange,
  disabled = false,
}: {
  q: Q;
  answer: Answer;
  onChange: (a: Answer) => void;
  api: EducationClient;
  disabled?: boolean;
}) {
  if (q.type === "single" || q.type === "multiple")
    return (
      <fieldset disabled={disabled}>
        <legend className="sr-only">Response to {q.stem}</legend>
        {q.options?.map((o) => (
          <label className="answer-option" key={o.id}>
            <input
              type={q.type === "single" ? "radio" : "checkbox"}
              name={"answer-" + q.id}
              checked={
                q.type === "single" && (answer.selection?.length ?? 0) > 1
                  ? false
                  : (answer.selection?.includes(o.id) ?? false)
              }
              onChange={(e) =>
                onChange({
                  ...answer,
                  selection:
                    q.type === "single"
                      ? [o.id]
                      : e.target.checked
                        ? [...(answer.selection ?? []), o.id]
                        : (answer.selection ?? []).filter((v) => v !== o.id),
                  visited: true,
                  invalid: false,
                })
              }
            />
            <span>
              {o.id}. {o.text}
              {o.asset_id && (
                <img
                  className="question-image"
                  src={api.url("assets/" + o.asset_id)}
                  alt={o.alt ?? ""}
                />
              )}
            </span>
          </label>
        ))}
        {!disabled && (
          <button
            type="button"
            onClick={() => onChange({ ...blank(), visited: true })}
          >
            Clear response
          </button>
        )}
      </fieldset>
    );
  if (q.type === "numerical")
    return (
      <Field name="Numerical response">
        <input
          disabled={disabled}
          inputMode="decimal"
          value={answer.number}
          onChange={(e) =>
            onChange({ ...answer, number: e.target.value, visited: true })
          }
        />
      </Field>
    );
  return (
    <Field name="Written response">
      <textarea
        rows={6}
        disabled={disabled}
        value={answer.text}
        onChange={(e) =>
          onChange({ ...answer, text: e.target.value, visited: true })
        }
      />
    </Field>
  );
}
function AssessmentWork({
  api,
  initial,
  staff,
  onBack,
}: {
  api: EducationClient;
  initial: Sitting;
  staff: boolean;
  onBack: () => void;
}) {
  const [sitting, setSitting] = useState(initial);
  const [answers, setAnswers] = useState(initial.answers);
  const [index, setIndex] = useState(0);
  const [status, setStatus] = useState("All recorded answers are saved.");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [manual, setManual] = useState(initial.manual_marks ?? {});
  const [reason, setReason] = useState("");
  const [remaining, setRemaining] = useState(0);
  const pending = useRef(false);
  const inflight = useRef(false);
  const operation = useRef<{
    id: string;
    answers: Record<string, Answer>;
    version: number;
  } | null>(null);
  const [serverCopy, setServerCopy] = useState<Sitting | null>(null);
  const latest = useRef(answers);
  latest.current = answers;
  const state = useRef(sitting);
  state.current = sitting;
  const questions = sitting.paper.spec.sections.flatMap((sec) => sec.questions);
  const q = questions[index];
  const active = sitting.state === "active" && !staff;
  const sync = useCallback(async () => {
    if (
      inflight.current ||
      !pending.current ||
      state.current.state !== "active"
    )
      return;
    inflight.current = true;
    setStatus("Saving answers…");
    try {
      if (!operation.current)
        operation.current = {
          id: crypto.randomUUID(),
          answers: latest.current,
          version: state.current.answer_version,
        };
      const snapshot = operation.current.answers;
      const next = await api.save(state.current.id, {
        operation_id: operation.current.id,
        version: operation.current.version,
        answers: snapshot,
      });
      setSitting(next);
      state.current = next;
      operation.current = null;
      if (snapshot === latest.current) {
        pending.current = false;
        setStatus(
          next.state === "active"
            ? "All recorded answers are saved."
            : "The sitting has ended. Saved responses are retained.",
        );
      } else setStatus("New changes waiting to save…");
      setError("");
    } catch (e) {
      setStatus("Unsaved changes retained in this open page.");
      setError(
        e instanceof Error
          ? e.message
          : "Connection interrupted. Retry saving before leaving.",
      );
    } finally {
      inflight.current = false;
    }
  }, [api]);
  useEffect(() => {
    if (!active) return;
    const timer = setInterval(() => void sync(), 1500);
    return () => clearInterval(timer);
  }, [answers, active, sync]);
  useEffect(() => {
    if (!active) return;
    const start = Date.now();
    const offset = new Date(initial.server_time).getTime() - start;
    const tick = setInterval(() => {
      const n = Math.max(
        0,
        new Date(state.current.deadline).getTime() - (Date.now() + offset),
      );
      setRemaining(Math.ceil(n / 1000));
      if (n === 0 && !inflight.current)
        void api
          .detail("sittings", state.current.id)
          .then((s) => {
            setSitting(s);
            state.current = s;
          })
          .catch((e) => setError(e.message));
    }, 1000);
    return () => clearInterval(tick);
  }, [active, api, initial.server_time]);
  useEffect(() => {
    if (!active) return;
    const warn = (e: BeforeUnloadEvent) => {
      if (pending.current) {
        e.preventDefault();
        e.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", warn);
    const online = () => void sync();
    window.addEventListener("online", online);
    return () => {
      window.removeEventListener("beforeunload", warn);
      window.removeEventListener("online", online);
    };
  }, [active, sync]);
  const run = async (fn: () => Promise<unknown>) => {
    setBusy(true);
    setError("");
    try {
      await fn();
    } catch (e) {
      setError(
        e instanceof Error ? e.message : "Operation interrupted; retry.",
      );
    } finally {
      setBusy(false);
    }
  };
  return (
    <section className="academic assessment-work">
      <div className="section-heading">
        <div>
          <h2>{sitting.paper.spec.title}</h2>
          <p>
            {sitting.state} · {sitting.paper.reference} ·{" "}
            {sitting.paper.maximum} marks
          </p>
        </div>
        <button
          disabled={busy || pending.current || inflight.current}
          onClick={onBack}
        >
          Back to work
        </button>
      </div>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {active && (
        <>
          <p className="exam-clock" role="timer">
            {Math.floor(remaining / 60)}:
            {String(remaining % 60).padStart(2, "0")} remaining · deadline{" "}
            {new Date(sitting.deadline).toLocaleString()}
          </p>
          <p role="status">{status}</p>
          <div className="actions">
            <button
              onClick={() =>
                void run(async () =>
                  setServerCopy(await api.detail("sittings", sitting.id)),
                )
              }
              disabled={busy}
            >
              Review saved version / resolve conflict
            </button>
            <button onClick={() => void sync()} disabled={busy}>
              Retry saving
            </button>
            <button
              onClick={() =>
                void run(async () => {
                  if (inflight.current)
                    throw new Error("Wait for the current save to finish.");
                  await sync();
                  if (pending.current)
                    throw new Error(
                      "Save or resolve the answer conflict before submitting.",
                    );
                  const s = await api.submit(sitting.id);
                  setSitting(s);
                })
              }
              disabled={busy}
            >
              Submit saved answers
            </button>
          </div>
          {serverCopy && (
            <section className="card">
              <h3>Compare saved work before merging</h3>
              <p>
                Your local answers are still in this page. Choose whether to
                keep them or use the server copy. A transferred writer must
                resume this assessment here before saving.
              </p>
              {questions.map((item) => (
                <p key={item.id}>
                  {item.stem}: saved{" "}
                  {JSON.stringify(serverCopy.answers[item.id!])} / local{" "}
                  {JSON.stringify(answers[item.id!])}
                </p>
              ))}
              <button
                onClick={() => {
                  setSitting(serverCopy);
                  state.current = serverCopy;
                  operation.current = null;
                  pending.current = true;
                  setServerCopy(null);
                  setStatus(
                    "Local responses retained for an explicit retry against the latest saved version.",
                  );
                }}
              >
                Keep local responses and retry against current version
              </button>
              <button
                onClick={() => {
                  setSitting(serverCopy);
                  state.current = serverCopy;
                  setAnswers(serverCopy.answers);
                  pending.current = false;
                  operation.current = null;
                  setServerCopy(null);
                  setStatus("Loaded the current saved responses.");
                }}
              >
                Use saved responses
              </button>
            </section>
          )}
          <p>{sitting.paper.spec.instructions}</p>
        </>
      )}
      <nav className="question-nav" aria-label="Assessment question navigation">
        {questions.map((item, i) => (
          <button
            key={item.id}
            aria-current={i === index ? "step" : undefined}
            onClick={() => setIndex(i)}
          >
            {i + 1}
            {answers[item.id!]?.review ? " ⚑" : ""}
            {answers[item.id!]?.selection?.length ||
            answers[item.id!]?.text ||
            answers[item.id!]?.number
              ? " •"
              : ""}
          </button>
        ))}
      </nav>
      {q && (
        <article className="card">
          <h3>
            Question {index + 1} · {q.marks} marks
          </h3>
          <QuestionContent q={q} api={api} options={false} />
          <AnswerControl
            api={api}
            q={q}
            answer={answers[q.id!] ?? blank()}
            disabled={!active}
            onChange={(a) => {
              pending.current = true;
              setAnswers({ ...answers, [q.id!]: a });
              setStatus("Changes waiting to save…");
            }}
          />
          {active && (
            <label className="check">
              <input
                type="checkbox"
                checked={answers[q.id!]?.review ?? false}
                onChange={(e) => {
                  pending.current = true;
                  setAnswers({
                    ...answers,
                    [q.id!]: {
                      ...(answers[q.id!] ?? blank()),
                      review: e.target.checked,
                    },
                  });
                }}
              />
              Flag for review
            </label>
          )}
          {sitting.result && (
            <>
              <p className="notice">
                {
                  sitting.result.items.find((i) => i.question_id === q.id)
                    ?.status
                }
                :{" "}
                {
                  sitting.result.items.find((i) => i.question_id === q.id)
                    ?.award
                }{" "}
                / {q.marks}
              </p>
              <p>
                {
                  sitting.result.items.find((i) => i.question_id === q.id)
                    ?.explanation
                }
              </p>
              <h4>Worked solution</h4>
              <p>{q.solution}</p>
              <p>Answer rule: {q.key?.join(", ") || answerRuleText(q)}</p>
            </>
          )}
          {staff &&
            ["short", "extended"].includes(q.type) &&
            sitting.state !== "active" && (
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  void run(async () =>
                    setSitting(
                      await api.mark(sitting.id, {
                        revision: sitting.revision,
                        reason,
                        manual_marks: manual,
                      }),
                    ),
                  );
                }}
              >
                <h4>Human rubric review</h4>
                {q.rubric?.map((c) => (
                  <Field
                    key={c.id}
                    name={`${c.description} / ${c.marks} maximum`}
                  >
                    <input
                      required
                      inputMode="decimal"
                      value={manual[q.id!]?.criteria?.[c.id] ?? ""}
                      onChange={(e) =>
                        setManual({
                          ...manual,
                          [q.id!]: {
                            criteria: {
                              ...(manual[q.id!]?.criteria ?? {}),
                              [c.id]: e.target.value,
                            },
                            feedback: manual[q.id!]?.feedback ?? "",
                            approved: false,
                          },
                        })
                      }
                    />
                  </Field>
                ))}
                <Field name="Marking feedback">
                  <textarea
                    required
                    value={manual[q.id!]?.feedback ?? ""}
                    onChange={(e) =>
                      setManual({
                        ...manual,
                        [q.id!]: {
                          criteria: manual[q.id!]?.criteria ?? {},
                          feedback: e.target.value,
                          approved: false,
                        },
                      })
                    }
                  />
                </Field>
                <Field name="Reason for result revision">
                  <textarea
                    required
                    minLength={5}
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                  />
                </Field>
                <button disabled={busy}>
                  Approve rubric marks and recalculate
                </button>
              </form>
            )}
          {!staff && sitting.state === "released" && (
            <TutorPanel
              key={q.id}
              api={api}
              sitting={sitting.id}
              question={q.id!}
            />
          )}
        </article>
      )}
      {sitting.result ? (
        <section className="card">
          <h3>Result evidence</h3>
          <ResultEvidencePanel
            key={sitting.revision}
            api={api}
            sitting={sitting}
            staff={staff}
          />
          <p className="result-score">
            {sitting.result.score} / {sitting.result.maximum}
          </p>
          <p>
            {sitting.result.correct} correct · {sitting.result.wrong} wrong ·{" "}
            {sitting.result.pending} pending human review ·{" "}
            {sitting.result.required} required contributions
          </p>
          <p className="muted">
            Optional uncounted answers remain visible and do not imply a
            learning gap. Marking revision {sitting.revision} ·{" "}
            {sitting.result.policy}
          </p>
          {staff && sitting.state === "marked" && (
            <button
              disabled={busy}
              onClick={() =>
                void run(async () => {
                  await api.release(sitting.id);
                  setSitting(await api.detail("sittings", sitting.id));
                })
              }
            >
              Release reviewed result
            </button>
          )}
        </section>
      ) : (
        sitting.state !== "active" && (
          <p className="notice">
            Your submission is retained. Results will appear after the teacher
            completes marking and releases them.
          </p>
        )
      )}
    </section>
  );
}
function TutorPanel({
  api,
  sitting,
  question,
}: {
  api: EducationClient;
  sitting: string;
  question: string;
}) {
  const [thread, setThread] = useState<A<"AcademicThread"> | null>(null);
  const [text, setText] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const refresh = useCallback(
    async () => setThread(await api.thread(sitting, question)),
    [api, sitting, question],
  );
  useEffect(() => {
    let active = true;
    api
      .thread(sitting, question)
      .then((t) => {
        if (active) setThread(t);
      })
      .catch((e) => {
        if (active) setError(e.message);
      });
    return () => {
      active = false;
    };
  }, [api, sitting, question]);
  const run = async (fn: () => Promise<unknown>) => {
    setBusy(true);
    setError("");
    try {
      await fn();
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Tutor unavailable. Your saved result is retained.",
      );
    } finally {
      setBusy(false);
    }
  };
  return (
    <section className="tutor">
      <h4>Ask about this released question</h4>
      <p className="muted">
        AI guidance is an interpretation, not additional evidence about your
        ability. Check it with the worked solution or your teacher.
      </p>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {thread?.messages.map((m, i) => (
        <article className="question-card" key={i}>
          <strong>{m.role === "user" ? "You" : "AI tutor"}</strong>
          <p className="question-stem">{m.text}</p>
        </article>
      ))}
      {thread?.job_id && (
        <p role="status">
          Reply job: {thread.job_state}. {thread.error}
        </p>
      )}
      <div className="actions">
        <button onClick={() => void run(refresh)} disabled={busy}>
          Refresh tutor reply
        </button>
        {thread?.job_id &&
          ["failed", "configuration_required", "cancelled"].includes(
            thread.job_state,
          ) && (
            <button
              disabled={busy}
              onClick={() =>
                void run(async () => {
                  await api.jobAction(thread.job_id, "retry");
                  await refresh();
                })
              }
            >
              Retry tutor job
            </button>
          )}
        {thread?.job_id && ["queued", "running"].includes(thread.job_state) && (
          <button
            disabled={busy}
            onClick={() =>
              void run(async () => {
                await api.jobAction(thread.job_id, "cancel");
                await refresh();
              })
            }
          >
            Cancel tutor job
          </button>
        )}
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void run(async () => {
            setThread(
              await api.ask({
                sitting_id: sitting,
                question_id: question,
                text,
              }),
            );
            setText("");
          });
        }}
      >
        <Field name="Your question for the tutor">
          <textarea
            required
            maxLength={4000}
            value={text}
            onChange={(e) => setText(e.target.value)}
          />
        </Field>
        <button
          disabled={
            busy || ["queued", "running"].includes(thread?.job_state ?? "")
          }
        >
          Ask contextual tutor
        </button>
      </form>
    </section>
  );
}
function ResultEvidencePanel({
  api,
  sitting,
  staff,
}: {
  api: EducationClient;
  sitting: Sitting;
  staff: boolean;
}) {
  const [topics, setTopics] = useState<A<"AcademicTopicEvidence">[]>([]);
  const [history, setHistory] = useState<A<"AcademicResultHistory">[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    Promise.all([
      sitting.state === "released"
        ? api.evidence(sitting.id)
        : Promise.resolve([]),
      staff ? api.history(sitting.id) : Promise.resolve([]),
    ])
      .then(([t, h]) => {
        if (active) {
          setTopics(t);
          setHistory(h);
        }
      })
      .catch((e) => {
        if (active) setError(e.message);
      });
    return () => {
      active = false;
    };
  }, [api, sitting.id, sitting.state, staff]);
  return (
    <>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {topics.length > 0 && (
        <details>
          <summary>Topic observations and evidence limits</summary>
          {topics.map((t) => (
            <article className="question-card" key={t.topic_id}>
              <strong>
                {t.name} · {t.status.replaceAll("_", " ")}
              </strong>
              <p>
                {t.correct} correct, {t.wrong} wrong/invalid, {t.required_blank}{" "}
                required blanks, {t.pending} pending, from {t.counted} counted
                opportunities.
              </p>
              <p className="muted">{t.explanation}</p>
            </article>
          ))}
        </details>
      )}
      {staff && (
        <details>
          <summary>Retained marking revisions</summary>
          {history.map((h) => (
            <article className="academic-row" key={h.revision}>
              <span>
                Revision {h.revision} · {h.result.score} / {h.result.maximum}
              </span>
              <span>
                {h.reason} · {h.created_at}
              </span>
            </article>
          ))}
        </details>
      )}
    </>
  );
}
