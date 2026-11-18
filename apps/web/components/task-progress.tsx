"use client";
import type { ReactNode } from "react";

export function TaskProgress({
  id,
  state,
  title,
  detail,
  progress,
  children,
}: {
  id?: string;
  state: string;
  title: string;
  detail: string;
  progress?: number;
  children?: ReactNode;
}) {
  const active = ["uploading", "queued", "running"].includes(state);
  const failed = ["failed", "configuration_required"].includes(state);
  const value =
    state === "completed"
      ? 100
      : progress && progress > 0
        ? Math.min(100, progress)
        : undefined;
  return (
    <section
      id={id}
      className={`task-progress ${active ? "is-active" : failed ? "needs-attention" : ""}`}
      aria-label={title}
    >
      <div className="task-progress-heading" role="status" aria-live="polite">
        <span
          className={active ? "task-spinner" : "task-symbol"}
          aria-hidden="true"
        >
          {active ? "" : failed ? "!" : state === "completed" ? "✓" : "—"}
        </span>
        <div>
          <span className="task-state">
            {state === "uploading"
              ? "Uploading"
              : state === "queued"
                ? "Queued"
                : state === "running"
                  ? "In progress"
                  : failed
                    ? "Needs attention"
                    : state === "completed"
                      ? "Ready"
                      : "Stopped"}
          </span>
          <h3>{title}</h3>
        </div>
      </div>
      <p>{detail}</p>
      {active && (
        <>
          <progress aria-label={`${title} progress`} max={100} value={value} />
          <p className="muted task-progress-caption">
            {value === undefined
              ? "Working… completion time is not yet known."
              : `${value}% of processing stages reported complete.`}{" "}
            Updates automatically.
          </p>
        </>
      )}
      {children && <div className="actions">{children}</div>}
    </section>
  );
}
