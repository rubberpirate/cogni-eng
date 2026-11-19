"use client";
import { useEffect, useState } from "react";
import { type A, type EducationClient } from "../../../clients/academic";
import { TaskProgress } from "./task-progress";
import { FormField as Field } from "./form-field";
import { StudyArtwork, WorkIcon } from "./workspace-art";
import { type LearningView } from "./learning-navigation";
type Paper = A<"AcademicPaperRecord">;
type Job = A<"AcademicJob">;
type Topic = A<"AcademicTopic">;
type Course = A<"AcademicCourse">;
export function LearningHome({
  navigate,
  papers,
  topics,
  jobs,
  openPaper,
}: {
  navigate: (view: LearningView) => void;
  openPaper: (paper: Paper) => void;
  papers: Paper[];
  topics: Topic[];
  jobs: Job[];
}) {
  const active = jobs.filter((j) =>
    ["queued", "running"].includes(j.state),
  ).length;
  const failed = jobs.filter((j) =>
    ["failed", "configuration_required"].includes(j.state),
  ).length;
  const published = papers.filter((p) => p.state === "published").length;
  return (
    <>
      <div className="studio-dashboard">
        <div className="studio-main">
          <section className="learning-welcome studio-welcome">
            <p className="eyebrow">Your teaching workspace</p>
            <h2>
              Good questions.
              <br />
              Clearer results.
            </h2>
            <p>
              Prepare a paper, set the answers and marks, then see how every
              student did.
            </p>
            <button className="primary" onClick={() => navigate("import")}>
              <WorkIcon name="import" /> Upload a question paper
            </button>
          </section>
          <section
            aria-labelledby="quick-actions-title"
            className="create-section"
          >
            <div className="section-heading">
              <h2 id="quick-actions-title">Start with questions</h2>
              <span className="muted">Choose what works for you</span>
            </div>
            <div className="creation-methods">
              {(
                [
                  [
                    "import",
                    "01",
                    "Upload a paper",
                    "PDF or photo → editable questions",
                    "OCR extracts questions and options.",
                  ],
                  [
                    "author",
                    "02",
                    "Enter manually",
                    "Write questions, choose answers",
                    "Add options and marks as you go.",
                  ],
                  [
                    "generate",
                    "03",
                    "Create with AI",
                    "Subject, topic and difficulty",
                    "Review generated drafts before use.",
                  ],
                ] as const
              ).map(([view, number, title, subtitle, detail]) => (
                <button
                  className={"creation-method method-" + view}
                  key={view}
                  onClick={() => navigate(view)}
                >
                  <span className="method-top">
                    <WorkIcon name={view === "author" ? "curriculum" : view} />
                    <span>{number}</span>
                  </span>
                  <strong>{title}</strong>
                  <span>{subtitle}</span>
                  <small>{detail}</small>
                  <span className="method-arrow" aria-hidden="true">
                    ↗
                  </span>
                </button>
              ))}
            </div>
          </section>
        </div>
        <aside className="studio-aside" aria-label="Workspace overview">
          <StudyArtwork compact />
          <h2>Your workspace</h2>
          <div className="workspace-metrics">
            <div>
              <strong>{papers.length}</strong>
              <span>Assessments</span>
            </div>
            <div>
              <strong>{topics.length}</strong>
              <span>Topics</span>
            </div>
          </div>
          <p className="muted">
            {published} published · {papers.length - published} drafts
          </p>
          <div className="next-task-list">
            <button onClick={() => navigate("papers")}>
              <WorkIcon name="papers" />
              <span>
                <strong>Answers & marking</strong>
                <small>Build and assign an assessment</small>
              </span>
              <span aria-hidden="true">↗</span>
            </button>
            <button onClick={() => navigate("scan")}>
              <WorkIcon name="scan" />
              <span>
                <strong>Grade OMR sheets</strong>
                <small>Individual or batch upload</small>
              </span>
              <span aria-hidden="true">↗</span>
            </button>
            <button onClick={() => navigate("results")}>
              <WorkIcon name="results" />
              <span>
                <strong>Student results</strong>
                <small>Marks and question breakdown</small>
              </span>
              <span aria-hidden="true">↗</span>
            </button>
          </div>
          <button className="text-action" onClick={() => navigate("jobs")}>
            {active} processing · {failed} need attention{" "}
            <WorkIcon name="arrow" />
          </button>
        </aside>
      </div>
      <section aria-labelledby="recent-assessments-title">
        <div className="section-heading assessment-heading">
          <div>
            <h2 id="recent-assessments-title">Your assessments</h2>
            <p className="muted">Pick up where you left off.</p>
          </div>
          <button onClick={() => navigate("papers")}>
            See assessments <WorkIcon name="arrow" />
          </button>
        </div>
        {papers.length ? (
          <div className="assessment-card-grid">
            {papers.slice(0, 4).map((p, i) => (
              <article className="assessment-tile" key={p.id}>
                <div className={"assessment-cover tone-" + i}>
                  <WorkIcon name="papers" />
                  <span className="state-pill">
                    {p.state === "draft" ? "Draft" : "Published"}
                  </span>
                  <span className="cover-lines" aria-hidden="true">
                    <i />
                    <i />
                    <i />
                  </span>
                </div>
                <div className="assessment-tile-body">
                  <h3>{p.spec.title}</h3>
                  <p className="muted">{p.reference}</p>
                  <div className="assessment-meta">
                    <span>
                      {p.spec.sections.reduce(
                        (n, section) => n + section.questions.length,
                        0,
                      )}{" "}
                      questions
                    </span>
                    <span>{p.maximum} marks</span>
                  </div>
                  <button onClick={() => openPaper(p)}>
                    Open assessment <WorkIcon name="arrow" />
                  </button>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <p className="empty-message">
            No assessments yet. Build your first test or try the sample kit
            below.
          </p>
        )}
      </section>
      <section className="sample-invitation">
        <span className="icon-tile">
          <WorkIcon name="samples" />
        </span>
        <div>
          <h2>A first run, without the preparation.</h2>
          <p>
            Try an original question paper, answer key and filled OMR sheets.
            Follow the complete flow with a test learner.
          </p>
        </div>
        <button onClick={() => navigate("samples")}>
          Try sample files <WorkIcon name="arrow" />
        </button>
      </section>
    </>
  );
}

export function GenerationFlow({
  api,
  topics,
  courses,
  onTopic,
  onReview,
  initialTopic,
  initialJob,
  onJob,
}: {
  initialJob?: Job;
  onJob: (job: Job) => void;
  initialTopic: string;
  api: EducationClient;
  topics: Topic[];
  courses: Course[];
  onTopic: (id: string) => void;
  onReview: (id: string) => void;
}) {
  const [selection, setSelection] = useState(initialTopic);
  const [createdTopic, setCreatedTopic] = useState<Topic | null>(null);
  const availableTopics =
    createdTopic && !topics.some((t) => t.id === createdTopic.id)
      ? [...topics, createdTopic]
      : topics;
  const [job, setJobState] = useState<Job | null>(initialJob ?? null);
  const setJob = (next: Job) => {
    setJobState(next);
    onJob(next);
  };
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const chosen = selection || availableTopics[0]?.id || "new";
  useEffect(() => {
    if (!job || !["queued", "running"].includes(job.state)) return;
    let active = true;
    let pending = false;
    const poll = async () => {
      if (pending) return;
      pending = true;
      try {
        const next = await api.detail("jobs", job.id);
        if (active) {
          setJob(next);
          setError("");
        }
      } catch (e) {
        if (active)
          setError(
            e instanceof Error
              ? e.message
              : "Connection lost. Your request is saved; refresh its status.",
          );
      } finally {
        pending = false;
      }
    };
    const timer = setInterval(() => void poll(), 3000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, job?.id, job?.state]);
  const unavailable =
    !!job && job.state === "failed" && /HTTP 50[234]/.test(job.error);
  const running = !!job && ["queued", "running"].includes(job.state);
  const labels: Record<string, string> = {
    queued: "Waiting to start",
    running: "Creating draft questions",
    completed: "Your draft questions are ready",
    failed: "The questions could not be generated",
    configuration_required: "The AI provider needs attention",
    cancelled: "Request cancelled",
  };
  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const f = new FormData(e.currentTarget);
    try {
      let id = chosen;
      if (chosen === "new") {
        const course = courses.find((c) => c.id === String(f.get("course")));
        if (!course)
          throw new Error(
            "Choose a course first. A coordinator can add one under People & organization.",
          );
        const t = await api.topics({
          id: "",
          campus_id: course.campus_id,
          course_id: course.id,
          subject: String(f.get("subject")),
          name: String(f.get("name")),
          outcome: String(f.get("outcome")),
          language: "en",
        });
        id = t.id;
        setCreatedTopic(t);
        setSelection(id);
      }
      onTopic(id);
      setJob(
        await api.generate({
          topic_id: id,
          count: Number(f.get("count")),
          difficulty: String(f.get("difficulty")),
        }),
      );
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Could not start generation. Your input is still here.",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="generation-layout">
      <section className="card">
        <h2>Tell us what to teach</h2>
        <p>
          Choose a topic or describe a new one. Generated questions are saved as
          drafts for your review.
        </p>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <form onSubmit={submit}>
          <fieldset disabled={busy || running}>
            <legend className="sr-only">Question generation</legend>
            <Field name="Topic to generate questions for">
              <select
                value={chosen}
                onChange={(e) => setSelection(e.target.value)}
              >
                {availableTopics.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.subject} / {t.name}
                  </option>
                ))}
                <option value="new">+ Describe a new topic</option>
              </select>
            </Field>
            {chosen === "new" ? (
              <>
                <Field name="Course">
                  <select name="course" required>
                    <option value="">Choose a course</option>
                    {courses.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field name="Subject">
                  <input name="subject" required defaultValue="Mathematics" />
                </Field>
                <Field name="Topic name">
                  <input
                    name="name"
                    required
                    minLength={2}
                    placeholder="For example, adding fractions"
                  />
                </Field>
                <Field name="What should learners be able to do?">
                  <textarea
                    name="outcome"
                    required
                    minLength={5}
                    placeholder="For example, add fractions with the same denominator and explain each step."
                  />
                </Field>
              </>
            ) : (
              <p className="topic-context">
                {availableTopics.find((t) => t.id === chosen)?.outcome}
              </p>
            )}
            <div className="academic-columns">
              <Field name="How many questions?">
                <input
                  name="count"
                  type="number"
                  min={1}
                  max={20}
                  defaultValue={5}
                  required
                />
              </Field>
              <Field name="Difficulty">
                <select name="difficulty">
                  <option value="introductory">Easy — introductory</option>
                  <option value="standard">Standard</option>
                  <option value="challenging">Challenging</option>
                </select>
              </Field>
            </div>
            <button className="primary" disabled={!courses.length}>
              {busy
                ? "Starting…"
                : running
                  ? "Generation in progress…"
                  : "Generate draft questions"}
            </button>
          </fieldset>
        </form>
      </section>
      <section className="card generation-status" aria-live="polite">
        {(busy || running) && (
          <TaskProgress
            state={busy && !running ? "uploading" : job!.state}
            title={
              busy && !running
                ? "Sending your generation request"
                : job!.state === "queued"
                  ? "Your request is queued"
                  : "Generating questions and answer proposals"
            }
            detail="Your topic is saved with the request. Drafts will need faculty review before use."
            progress={running ? job?.progress : undefined}
          />
        )}
        <p className="eyebrow">Request status</p>
        <h2>
          {job
            ? unavailable
              ? "The AI service is temporarily unavailable"
              : (labels[job.state] ?? job.state)
            : "Your questions will appear here"}
        </h2>
        {!job ? (
          <>
            <p>
              No request has been sent yet. Your institution's configured AI
              provider will receive the topic and learning outcome.
            </p>
            <ol>
              <li>Describe the learning goal.</li>
              <li>Wait for the saved request to finish.</li>
              <li>Check each question, answer and explanation.</li>
            </ol>
          </>
        ) : (
          <>
            <p>
              {running
                ? "You can leave this screen. The request is saved and continues in the background. Status updates every few seconds."
                : job.state === "completed"
                  ? "Review every answer key before approving questions for a test."
                  : "Your topic and request are saved. Existing questions and assessments are unaffected."}
            </p>
            {job.error &&
              (unavailable ? (
                <div className="provider-recovery">
                  <p>
                    Gemini could not complete this request. Wait a little, then
                    retry the saved request below.
                  </p>
                  <p>
                    You can keep working with your question bank or try the
                    sample paper while the service recovers.
                  </p>
                  <details>
                    <summary>Technical details</summary>
                    <p className="muted">{job.error}</p>
                  </details>
                </div>
              ) : (
                <p className="error">{job.error}</p>
              ))}
            <p>
              <a href="#education/samples">Try ready-made sample questions</a> ·{" "}
              <a href="#education/curriculum">Write questions yourself</a>
            </p>
            <p className="muted">
              {job.provider
                ? `Provider: ${job.provider} ${job.model}`
                : "Provider details appear after a response is received."}
            </p>
            <div className="actions">
              {job.state === "completed" && (
                <button
                  className="primary"
                  onClick={() => onReview(job.resource_id)}
                >
                  Review generated questions
                </button>
              )}
              {["failed", "configuration_required", "cancelled"].includes(
                job.state,
              ) && (
                <button
                  disabled={busy}
                  onClick={async () => {
                    setBusy(true);
                    try {
                      await api.jobAction(job.id, "retry");
                      setJob(await api.detail("jobs", job.id));
                    } catch (e) {
                      setError(e instanceof Error ? e.message : "Retry failed");
                    } finally {
                      setBusy(false);
                    }
                  }}
                >
                  Retry this request
                </button>
              )}
              {running && (
                <button
                  disabled={busy}
                  onClick={async () => {
                    try {
                      await api.jobAction(job.id, "cancel");
                      setJob(await api.detail("jobs", job.id));
                    } catch (e) {
                      setError(
                        e instanceof Error ? e.message : "Could not cancel",
                      );
                    }
                  }}
                >
                  Cancel request
                </button>
              )}
              <button
                onClick={() =>
                  void api
                    .detail("jobs", job.id)
                    .then(setJob)
                    .catch((e) => setError(e.message))
                }
              >
                Check status now
              </button>
            </div>
            {job.provider === "development_fixture" && (
              <p className="environment">
                Demonstration output. This was not generated by an external AI
                service.
              </p>
            )}
          </>
        )}
      </section>
    </div>
  );
}
export function SamplesFlow({
  api,
  courses,
  onOpen,
}: {
  api: EducationClient;
  courses: Course[];
  onOpen: (kit: A<"AcademicSampleKit">) => void;
}) {
  const [course, setCourse] = useState("");
  const [learners, setLearners] = useState<A<"AcademicRecipient">[]>([]);
  const [learner, setLearner] = useState("");
  const [kit, setKit] = useState<A<"AcademicSampleKit"> | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    setLearner("");
    setLearners([]);
    if (course)
      api
        .courseLearners(course)
        .then((l) => {
          if (active) setLearners(l);
        })
        .catch((e) => {
          if (active) setError(e.message);
        });
    return () => {
      active = false;
    };
  }, [api, course]);
  return (
    <>
      <section className="sample-intro">
        <p className="eyebrow">Original, ready-to-use test material</p>
        <h2>Try the full flow in a few minutes.</h2>
        <p>
          Create a practice test for a test learner. We will give you matching
          files, including filled answer sheets with known scores. No AI
          provider is needed.
        </p>
        <div className="sample-downloads">
          <a
            href="/samples/question-paper.pdf"
            target="_blank"
            rel="noreferrer"
          >
            Preview the question paper ↗
          </a>
          <a href="/samples/marking-guide.pdf" target="_blank" rel="noreferrer">
            See the answer key ↗
          </a>
          <a href="/samples/questions.csv" download>
            Download questions CSV
          </a>
          <a href="/samples/questions.json" download>
            Download questions JSON
          </a>
        </div>
      </section>
      <div className="academic-columns">
        <section className="card">
          <h2>1. Prepare your sample test</h2>
          <p>
            This creates a new six-question practice assessment and assigns it
            to the learner below. Choose a test account.
          </p>
          {error && (
            <p className="error" role="alert">
              {error}
            </p>
          )}
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              setBusy(true);
              setError("");
              try {
                setKit(
                  await api.sampleKit({
                    course_id: course,
                    learner_id: learner,
                  }),
                );
              } catch (e) {
                setError(
                  e instanceof Error
                    ? e.message
                    : "Could not prepare samples. Please retry.",
                );
              } finally {
                setBusy(false);
              }
            }}
          >
            <Field name="Sample course">
              <select
                value={course}
                required
                onChange={(e) => {
                  setCourse(e.target.value);
                  setKit(null);
                }}
              >
                <option value="">Choose a course</option>
                {courses.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field name="Test learner">
              <select
                value={learner}
                required
                disabled={!course}
                onChange={(e) => setLearner(e.target.value)}
              >
                <option value="">Choose a test learner</option>
                {learners.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.name} / {l.reference}
                  </option>
                ))}
              </select>
            </Field>
            {course && learners.length === 0 && (
              <p className="muted">
                This course has no eligible active learners. Add an enrollment
                under People & organization, or choose the manual test course.
              </p>
            )}
            <button className="primary" disabled={busy || !learner}>
              {busy ? "Preparing files…" : "Create a sample practice test"}
            </button>
          </form>
          {kit && (
            <div className="notice" role="status">
              <strong>Your matching files are ready.</strong>
              <p>The sheet identity belongs to this new practice test.</p>
              <a
                className="button-link"
                href={api.url(
                  `sample-files/${kit.paper_id}?assignment=${kit.assignment_id}`,
                )}
                download
              >
                Download complete test kit (.zip)
              </a>
              <button onClick={() => onOpen(kit)}>Go to Scan & check</button>
            </div>
          )}
        </section>
        <section className="card">
          <h2>2. Upload one filled sheet</h2>
          <ol className="sample-instructions">
            <li>
              <strong>Download and unzip the kit.</strong>
              <p>
                Start with <code>omr-correct.png</code>. The PDF version tests
                scanned-PDF intake.
              </p>
            </li>
            <li>
              <strong>Open Scan & check.</strong>
              <p>
                Choose the new sample test, upload the sheet and wait for
                processing.
              </p>
            </li>
            <li>
              <strong>Check identity and save the result.</strong>
              <p>
                The correctly filled sheet scores 12 / 12. The mixed-answer
                sheet scores 5 / 12.
              </p>
            </li>
          </ol>
          <details>
            <summary>What else can I test?</summary>
            <p>
              <code>omr-needs-review.png</code> has two marks on question 1:
              correct it to B before committing.{" "}
              <code>duplicate-pages.pdf</code> repeats the same page to test
              deduplication. Print <code>omr-blank.pdf</code> to test a real
              scanner or camera.
            </p>
            <p>
              Create a fresh sample test before committing a different example:
              each learner assignment keeps one sitting.
            </p>
          </details>
          <p className="muted">
            Original CC0-1.0 demonstration content. These registered sheets are
            compatible with this reader; unrelated commercial OMR layouts are
            not automatically supported.
          </p>
        </section>
      </div>
    </>
  );
}
export function ResultsFlow({
  api,
  papers,
  onSitting,
  initialPaper,
}: {
  initialPaper: string;
  api: EducationClient;
  papers: Paper[];
  onSitting: (s: A<"AcademicSitting">) => void;
}) {
  const [paper, setPaper] = useState(initialPaper);
  useEffect(() => setPaper(initialPaper), [initialPaper]);
  const [rows, setRows] = useState<A<"AcademicAssigned">[]>([]);
  const [sittings, setSittings] = useState<
    Record<string, A<"AcademicSitting">>
  >({});
  const [unavailable, setUnavailable] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [refresh, setRefresh] = useState(0);
  useEffect(() => {
    let active = true;
    setRows([]);
    setSittings({});
    setUnavailable([]);
    setError("");
    if (!paper) {
      setLoading(false);
      return;
    }
    setLoading(true);
    void (async () => {
      try {
        const assigned = await api.list("assignments", { paper });
        if (!active) return;
        setRows(assigned);
        const submitted = assigned.filter((r) => r.sitting_id);
        for (let i = 0; i < submitted.length && active; i += 6) {
          const group = submitted.slice(i, i + 6);
          const responses = await Promise.allSettled(
            group.map((r) => api.detail("sittings", r.sitting_id)),
          );
          if (!active) return;
          responses.forEach((result, j) => {
            const id = group[j].sitting_id;
            if (result.status === "fulfilled")
              setSittings((current) => ({ ...current, [id]: result.value }));
            else setUnavailable((current) => [...current, id]);
          });
        }
      } catch (e) {
        if (active)
          setError(
            e instanceof Error ? e.message : "Could not load student results.",
          );
      } finally {
        if (active) setLoading(false);
      }
    })();
    return () => {
      active = false;
    };
  }, [api, paper, refresh]);
  const filtered = rows.filter((r) =>
    r.learner.toLowerCase().includes(query.toLowerCase()),
  );
  const graded = Object.values(sittings).filter((s) => s.result).length;
  return (
    <section className="card student-results" aria-busy={loading}>
      <div className="section-heading">
        <div>
          <h2>Student results</h2>
          <p className="muted">Every mark, with the answers behind it.</p>
        </div>
        <button
          disabled={loading || !paper}
          onClick={() => setRefresh((n) => n + 1)}
        >
          Refresh results
        </button>
      </div>
      <div className="results-toolbar">
        <Field name="Assessment to review">
          <select value={paper} onChange={(e) => setPaper(e.target.value)}>
            <option value="">Choose an assessment</option>
            {papers
              .filter((p) => p.state === "published")
              .map((p) => (
                <option key={p.id} value={p.id}>
                  {p.spec.title}
                </option>
              ))}
          </select>
        </Field>
        <Field name="Search students">
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Student name"
          />
        </Field>
      </div>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {!paper ? (
        <p className="empty-message">
          Choose an assessment to see who has submitted and what needs review.
        </p>
      ) : (
        <>
          <div className="results-summary" role="status">
            <span>
              <strong>{rows.length}</strong> students
            </span>
            <span>
              <strong>{graded}</strong> graded
            </span>
            <span>
              <strong>{rows.filter((r) => !r.sitting_id).length}</strong> not
              started
            </span>
            {loading && <span>Loading results…</span>}
          </div>
          {!loading && rows.length === 0 && !error ? (
            <p className="empty-message">
              No learners assigned to this assessment yet. Assign it from
              Assessments.
            </p>
          ) : (
            <div
              className="results-table-wrap"
              role="region"
              aria-label="Student marks and question breakdown"
              tabIndex={0}
            >
              <table className="results-table">
                <thead>
                  <tr>
                    <th scope="col">Student</th>
                    <th scope="col">Status</th>
                    <th scope="col" className="numeric">
                      Marks
                    </th>
                    <th scope="col" className="numeric">
                      Correct / wrong
                    </th>
                    <th scope="col">Correct questions</th>
                    <th scope="col">Review</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((r) => {
                    const sitting = sittings[r.sitting_id];
                    const result = sitting?.result;
                    const order =
                      sitting?.paper.spec.sections
                        .flatMap((s) => s.questions)
                        .map((q) => q.id) ?? [];
                    const correct =
                      result?.items
                        .filter((i) => i.status === "correct")
                        .map((i) => order.indexOf(i.question_id) + 1)
                        .filter((n) => n > 0)
                        .sort((a, b) => a - b) ?? [];
                    return (
                      <tr key={r.id}>
                        <th scope="row">
                          <span className="student-avatar" aria-hidden="true">
                            {r.learner.trim().slice(0, 1)}
                          </span>
                          {r.learner}
                        </th>
                        <td>
                          <span className="state-pill">
                            {r.state.replaceAll("_", " ")}
                          </span>
                        </td>
                        <td className="numeric">
                          {result ? (
                            <>
                              <strong>{result.score}</strong>
                              <span className="muted"> / {result.maximum}</span>
                              {result.pending > 0 && (
                                <small>
                                  Provisional · {result.pending} awaiting review
                                </small>
                              )}
                            </>
                          ) : unavailable.includes(r.sitting_id) ? (
                            "Unavailable"
                          ) : r.sitting_id && loading ? (
                            "Loading…"
                          ) : (
                            "—"
                          )}
                        </td>
                        <td className="numeric">
                          {result ? `${result.correct} / ${result.wrong}` : "—"}
                        </td>
                        <td className="correct-question-list">
                          {result
                            ? correct.length
                              ? correct.join(", ")
                              : "None"
                            : "—"}
                        </td>
                        <td>
                          {sitting ? (
                            <button onClick={() => onSitting(sitting)}>
                              {r.state === "awaiting_review"
                                ? "Mark written answers"
                                : "Open result & answers"}
                            </button>
                          ) : r.sitting_id ? (
                            <button
                              onClick={() => setRefresh((n) => n + 1)}
                              disabled={loading}
                            >
                              Retry result
                            </button>
                          ) : (
                            <span className="muted">Not started</span>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              {!loading && filtered.length === 0 && rows.length > 0 && (
                <p className="empty-message">No students match your search.</p>
              )}
            </div>
          )}
        </>
      )}
    </section>
  );
}
