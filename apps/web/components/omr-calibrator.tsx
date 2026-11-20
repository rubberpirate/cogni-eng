"use client";
import { useMemo, useRef, useState, type PointerEvent } from "react";
import type { A } from "../../../clients/academic";
import {
  blockRect,
  buildGrid,
  clamp,
  fitRect,
  gridRows,
  setBlockRect,
  type Calibration,
  type Point,
  type Rect,
} from "../../../clients/omr-calibration";
import { FormField as Field } from "./form-field";

type Selection = { kind: "area" | "block" | "question"; id: number };
type Gesture = {
  start: Point;
  original: Rect;
  selection: Selection;
  action: string;
  pointer: number;
  rect?: Rect;
};
const handles = ["nw", "ne", "sw", "se"] as const;

export function OMRCalibrator({
  url,
  calibration: c,
  onChange,
  reading,
  cornerTarget,
  onCorner,
  onCancelCorner,
}: {
  url: string;
  calibration: Calibration;
  onChange: (c: Calibration) => void;
  reading: A<"AcademicGridReading"> | null;
  cornerTarget: string | null;
  onCorner: (p: Point) => void;
  onCancelCorner: () => void;
}) {
  const [aspect, setAspect] = useState(1);
  const [pixels, setPixels] = useState({ width: 1000, height: 1000 });
  const [mode, setMode] = useState<"select" | Selection["kind"]>("area");
  const [selected, setSelected] = useState<Selection | null>(null);
  const [total, setTotal] = useState(40);
  const [columns, setColumns] = useState(1);
  const [first, setFirst] = useState(1);
  const [zoom, setZoom] = useState(100);
  const [showRows, setShowRows] = useState(true);
  const [draft, setDraft] = useState<Rect | null>(null);
  const [notice, setNotice] = useState(
    "Drag across the answer bubbles to select the full answer area.",
  );
  const [history, setHistory] = useState<Calibration[]>([]);
  const gesture = useRef<Gesture | null>(null);
  const svg = useRef<SVGSVGElement | null>(null);
  const rows = useMemo(() => gridRows(c, aspect), [c, aspect]);
  const results = useMemo(
    () => new Map(reading?.rows.map((r) => [r.number, r])),
    [reading],
  );
  const customFrames = useMemo(
    () => new Set(c.frames?.map((f) => f.number)),
    [c.frames],
  );
  const frameFor = (s: Selection) =>
    s.kind === "area"
      ? c.area
      : s.kind === "block"
        ? c.blocks[s.id]
          ? blockRect(c, s.id, aspect)
          : null
        : rows.find((r) => r.number === s.id)?.rect;
  const selectedRect = selected ? frameFor(selected) : null;

  function commit(next: Calibration) {
    setHistory((old) => [...old.slice(-19), c]);
    onChange(next);
  }
  function apply(s: Selection, rect: Rect, createArea = false) {
    if (
      rect.width < 2 * c.radius + 0.001 ||
      rect.height < (2 * c.radius) / aspect
    ) {
      setNotice("Draw a larger frame so the sampling circles fit inside it.");
      return;
    }
    let next = c;
    if (s.kind === "area") {
      if (createArea) {
        if (
          total < 1 ||
          total > 500 ||
          columns < 1 ||
          columns > 8 ||
          Math.ceil(total / columns) > 200 ||
          first < 1 ||
          !Number.isInteger(total + columns + first)
        ) {
          setNotice(
            "Choose 1–500 questions, 1–8 columns and at most 200 questions per column.",
          );
          return;
        }
        next = buildGrid(c, rect, total, columns, first, aspect);
      } else next = { ...c, area: rect };
    } else if (s.kind === "block") next = setBlockRect(c, s.id, rect, aspect);
    else
      next = {
        ...c,
        frames: [
          ...(c.frames ?? []).filter((f) => f.number !== s.id),
          { number: s.id, rect },
        ],
      };
    if (
      next.area &&
      s.kind !== "area" &&
      (rect.x < next.area.x - 1e-8 ||
        rect.y < next.area.y - 1e-8 ||
        rect.x + rect.width > next.area.x + next.area.width + 1e-8 ||
        rect.y + rect.height > next.area.y + next.area.height + 1e-8)
    ) {
      setNotice(
        "Keep this frame inside the answer area, or resize the answer area first.",
      );
      return;
    }
    commit(next);
    setSelected(s);
    setNotice(
      s.kind === "area" && createArea
        ? `Created ${total} question frames. Redraw each column if its bubbles need alignment.`
        : `${s.kind === "question" ? `Question ${s.id}` : s.kind === "block" ? `Block ${s.id + 1}` : "Answer area"} frame updated.`,
    );
  }
  function point(e: PointerEvent<SVGSVGElement>): Point {
    const bounds = e.currentTarget.getBoundingClientRect();
    return {
      x: clamp((e.clientX - bounds.left) / bounds.width),
      y: clamp((e.clientY - bounds.top) / bounds.height),
    };
  }
  function begin(e: PointerEvent<SVGSVGElement>) {
    if (e.button !== 0) return;
    const p = point(e);
    if (cornerTarget) {
      onCorner(p);
      return;
    }
    const hit = (e.target as Element).closest<SVGElement>("[data-frame-kind]");
    const handle = (e.target as SVGElement).dataset.handle;
    let selection: Selection;
    let action = "draw";
    if (handle && selected) {
      selection = selected;
      action = handle;
    } else if (mode === "select") {
      if (!hit) {
        setSelected(null);
        return;
      }
      selection = {
        kind: hit.dataset.frameKind as Selection["kind"],
        id: Number(hit.dataset.frameId),
      };
      action = "move";
    } else {
      selection = {
        kind: mode,
        id:
          mode === "area"
            ? 0
            : selected?.kind === mode
              ? selected.id
              : mode === "block"
                ? 0
                : (rows[0]?.number ?? 1),
      };
    }
    const original =
      action === "draw"
        ? { x: p.x, y: p.y, width: 0, height: 0 }
        : frameFor(selection);
    if (!original) return;
    e.preventDefault();
    e.currentTarget.focus();
    e.currentTarget.setPointerCapture(e.pointerId);
    gesture.current = {
      start: p,
      original,
      selection,
      action,
      pointer: e.pointerId,
    };
    setSelected(selection);
    setDraft(original);
  }
  function move(e: PointerEvent<SVGSVGElement>) {
    const g = gesture.current;
    if (!g || g.pointer !== e.pointerId) return;
    const p = point(e),
      dx = p.x - g.start.x,
      dy = p.y - g.start.y;
    let rect: Rect;
    if (g.action === "draw")
      rect = {
        x: Math.min(g.start.x, p.x),
        y: Math.min(g.start.y, p.y),
        width: Math.abs(dx),
        height: Math.abs(dy),
      };
    else if (g.action === "move")
      rect = fitRect({
        ...g.original,
        x: g.original.x + dx,
        y: g.original.y + dy,
      });
    else {
      const left = g.action.includes("w")
        ? clamp(g.original.x + dx, 0, g.original.x + g.original.width - 0.001)
        : g.original.x;
      const top = g.action.includes("n")
        ? clamp(g.original.y + dy, 0, g.original.y + g.original.height - 0.001)
        : g.original.y;
      const right = g.action.includes("e")
        ? clamp(g.original.x + g.original.width + dx, left + 0.001, 1)
        : g.original.x + g.original.width;
      const bottom = g.action.includes("s")
        ? clamp(g.original.y + g.original.height + dy, top + 0.001, 1)
        : g.original.y + g.original.height;
      rect = { x: left, y: top, width: right - left, height: bottom - top };
    }
    g.rect = rect;
    setDraft(rect);
  }
  function cancel() {
    const g = gesture.current;
    gesture.current = null;
    setDraft(null);
    if (g && svg.current?.hasPointerCapture(g.pointer))
      svg.current.releasePointerCapture(g.pointer);
    onCancelCorner();
    setNotice("Selection cancelled. Saved frames are unchanged.");
  }
  function end(e: PointerEvent<SVGSVGElement>) {
    const g = gesture.current;
    if (!g) return;
    gesture.current = null;
    if (e.currentTarget.hasPointerCapture(g.pointer))
      e.currentTarget.releasePointerCapture(g.pointer);
    const rect = g.rect;
    if (rect && rect.width > 0.001 && rect.height > 0.001)
      apply(
        g.selection,
        rect,
        g.action === "draw" && g.selection.kind === "area",
      );
    setDraft(null);
  }
  function download() {
    const href = URL.createObjectURL(
      new Blob([JSON.stringify(c, null, 2)], { type: "application/json" }),
    );
    const a = document.createElement("a");
    a.href = href;
    a.download = "omr-calibration.json";
    a.click();
    setTimeout(() => URL.revokeObjectURL(href), 1000);
  }
  return (
    <section className="omr-calibrator" aria-label="Visual OMR calibration">
      <div className="section-heading">
        <div>
          <h3>Frame your answers</h3>
          <p className="muted">
            Draw around the bubbles, leaving question numbers outside. Use a
            question frame to fix an individual row.
          </p>
        </div>
      </div>
      <div className="omr-grid-setup">
        <Field name="Total questions">
          <input
            type="number"
            min={1}
            max={500}
            value={total}
            onChange={(e) => setTotal(Number(e.target.value))}
          />
        </Field>
        <Field name="Answer columns">
          <input
            type="number"
            min={1}
            max={8}
            value={columns}
            onChange={(e) => setColumns(Number(e.target.value))}
          />
        </Field>
        <Field name="First question number">
          <input
            type="number"
            min={1}
            value={first}
            onChange={(e) => setFirst(Number(e.target.value))}
          />
        </Field>
      </div>
      <div
        className="omr-draw-toolbar"
        role="group"
        aria-label="Calibration tools"
      >
        {(
          [
            ["area", "Draw answer area"],
            ["block", "Draw answer block"],
            ["question", "Draw question frame"],
            ["select", "Move / resize"],
          ] as const
        ).map(([key, label]) => (
          <button
            key={key}
            type="button"
            className={mode === key ? "primary" : ""}
            aria-pressed={mode === key}
            onClick={() => {
              onCancelCorner();
              setMode(key);
              setNotice(
                key === "area"
                  ? "Drag a rectangle around the complete answer area. Drawing it builds a new grid."
                  : key === "block"
                    ? "Choose a block, then draw around its first and last bubble rows."
                    : key === "question"
                      ? "Choose a question, then draw tightly around its answer bubbles."
                      : "Drag a frame to move it. Drag a corner handle to resize.",
              );
            }}
          >
            {label}
          </button>
        ))}
        <button
          type="button"
          disabled={!history.length}
          onClick={() => {
            const prev = history.at(-1);
            if (prev) {
              onChange(prev);
              setHistory(history.slice(0, -1));
              setSelected(null);
            }
          }}
        >
          Undo frame change
        </button>
        <button type="button" onClick={download}>
          Save calibration
        </button>
        <label className="omr-load-calibration">
          Load calibration
          <input
            type="file"
            accept="application/json,.json"
            aria-label="Load OMR calibration"
            onChange={async (e) => {
              const file = e.target.files?.[0];
              if (!file) return;
              try {
                if (file.size > 256000)
                  throw Error("Calibration file is too large.");
                const raw = JSON.parse(await file.text()),
                  next = raw.calibration ?? raw;
                const validPoint = (p: Point) =>
                  p &&
                  Number.isFinite(p.x) &&
                  Number.isFinite(p.y) &&
                  p.x >= 0 &&
                  p.y >= 0 &&
                  p.x <= 1 &&
                  p.y <= 1;
                const validRect = (r: Rect) =>
                  r &&
                  [r.x, r.y, r.width, r.height].every(Number.isFinite) &&
                  r.x >= 0 &&
                  r.y >= 0 &&
                  r.width > 0 &&
                  r.height > 0 &&
                  r.x + r.width <= 1 &&
                  r.y + r.height <= 1;
                if (
                  !next ||
                  !Number.isInteger(next.options) ||
                  next.options < 2 ||
                  next.options > 8 ||
                  !Number.isFinite(next.radius) ||
                  next.radius < 0.001 ||
                  next.radius > 0.03 ||
                  !Array.isArray(next.blocks) ||
                  !next.blocks.length ||
                  next.blocks.length > 8 ||
                  next.blocks.some(
                    (b: Calibration["blocks"][number]) =>
                      !b ||
                      !Number.isInteger(b.first) ||
                      b.first < 1 ||
                      !Number.isInteger(b.count) ||
                      b.count < 1 ||
                      b.count > 200 ||
                      ![
                        b.top_left,
                        b.top_right,
                        b.bottom_left,
                        b.bottom_right,
                      ].every(validPoint),
                  ) ||
                  (next.area && !validRect(next.area)) ||
                  (next.frames &&
                    (!Array.isArray(next.frames) ||
                      next.frames.length > 500 ||
                      next.frames.some(
                        (f: A<"AcademicGridFrame">) =>
                          !f ||
                          !Number.isInteger(f.number) ||
                          f.number < 1 ||
                          !validRect(f.rect),
                      )))
                )
                  throw Error("Use a valid saved OMR calibration JSON file.");
                commit(next);
                setTotal(next.blocks.reduce((sum:number,b:Calibration['blocks'][number])=>sum+b.count,0));
                setColumns(next.blocks.length);
                setFirst(Math.min(...next.blocks.map((b:Calibration['blocks'][number])=>b.first)));
                setSelected(null);
                setNotice(
                  "Calibration loaded. Check the overlay against this sheet before reading.",
                );
              } catch (err) {
                setNotice(
                  err instanceof Error
                    ? err.message
                    : "Could not load calibration.",
                );
              }
              e.target.value = "";
            }}
          />
        </label>
      </div>
      <div className="omr-canvas-controls">
        <Field name="Selected answer block">
          <select
            value={selected?.kind === "block" ? selected.id : 0}
            onChange={(e) => {
              setSelected({ kind: "block", id: Number(e.target.value) });
              setMode("block");
            }}
          >
            {c.blocks.map((b, i) => (
              <option key={i} value={i}>
                Block {i + 1} · {b.first}–{b.first + b.count - 1}
              </option>
            ))}
          </select>
        </Field>
        <Field name="Selected question frame">
          <select
            value={
              selected?.kind === "question"
                ? selected.id
                : (rows[0]?.number ?? "")
            }
            onChange={(e) => {
              setSelected({ kind: "question", id: Number(e.target.value) });
              setMode("question");
            }}
          >
            {rows.map((r) => (
              <option key={r.number} value={r.number}>
                Question {r.number}
              </option>
            ))}
          </select>
        </Field>
        <Field name="Calibration zoom">
          <select
            value={zoom}
            onChange={(e) => setZoom(Number(e.target.value))}
          >
            {[50, 75, 100, 125, 150, 200, 250].map((n) => (
              <option key={n} value={n}>
                {n}%
              </option>
            ))}
          </select>
        </Field>
        <label className="check">
          <input
            type="checkbox"
            checked={showRows}
            onChange={(e) => setShowRows(e.target.checked)}
          />
          Question frames
        </label>
      </div>
      <p className="omr-canvas-status" role="status">
        {cornerTarget ?? notice}
      </p>
      <div className="omr-canvas-viewport">
        <div className="omr-canvas-image" style={{ width: `${zoom}%` }}>
          {/* Local image stays private until the existing authorized read action. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={url}
            alt="OMR sheet for visual calibration"
            draggable={false}
            onLoad={(e) => {
              const img = e.currentTarget;
              setAspect(img.naturalHeight / img.naturalWidth);
              setPixels({ width: img.naturalWidth, height: img.naturalHeight });
            }}
          />
          <svg
            ref={svg}
            viewBox="0 0 1 1"
            preserveAspectRatio="none"
            role="group"
            tabIndex={0}
            aria-label="OMR calibration canvas"
            className={mode === "select" ? "omr-select-tool" : "omr-draw-tool"}
            onPointerDown={begin}
            onPointerMove={move}
            onPointerUp={end}
            onPointerCancel={cancel}
            onKeyDown={(e) => {
              if (e.key === "Escape") {
                e.preventDefault();
                cancel();
              } else if (
                selected &&
                selectedRect &&
                ["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"].includes(
                  e.key,
                )
              ) {
                e.preventDefault();
                const step = e.shiftKey ? 10 : 1;
                apply(
                  selected,
                  fitRect({
                    ...selectedRect,
                    x:
                      selectedRect.x +
                      (e.key === "ArrowRight"
                        ? step
                        : e.key === "ArrowLeft"
                          ? -step
                          : 0) /
                        pixels.width,
                    y:
                      selectedRect.y +
                      (e.key === "ArrowDown"
                        ? step
                        : e.key === "ArrowUp"
                          ? -step
                          : 0) /
                        pixels.height,
                  }),
                );
              }
            }}
          >
            {c.area && (
              <path
                d={`M0 0H1V1H0Z M${c.area.x} ${c.area.y}v${c.area.height}h${c.area.width}v-${c.area.height}Z`}
                fillRule="evenodd"
                className="omr-outside-area"
              />
            )}
            {c.area && (
              <rect
                {...c.area}
                className="omr-area-frame"
                data-frame-kind="area"
                data-frame-id="0"
              />
            )}
            {c.blocks.map((b, i) => (
              <rect
                key={i}
                {...blockRect(c, i, aspect)}
                className="omr-block-frame"
                data-frame-kind="block"
                data-frame-id={i}
              >
                <title>
                  Block {i + 1}: {b.first}–{b.first + b.count - 1}
                </title>
              </rect>
            ))}
            {rows.map((r) => (
              <g key={r.number}>
                {showRows && (
                  <rect
                    {...r.rect}
                    className={
                      customFrames.has(r.number)
                        ? "omr-row-frame omr-custom-frame"
                        : "omr-row-frame"
                    }
                    data-frame-kind="question"
                    data-frame-id={r.number}
                  >
                    <title>Question {r.number}</title>
                  </rect>
                )}
                {r.centers.map((p, i) => {
                  const result = !draft ? results.get(r.number) : null;
                  return (
                    <ellipse
                      key={i}
                      cx={p.x}
                      cy={p.y}
                      rx={c.radius}
                      ry={c.radius / aspect}
                      className={
                        result?.selection.includes(i + 1)
                          ? "omr-bubble omr-filled"
                          : result?.uncertain
                            ? "omr-bubble omr-uncertain"
                            : "omr-bubble"
                      }
                    />
                  );
                })}
              </g>
            ))}
            {(draft ?? selectedRect) && (
              <g className="omr-selected-frame">
                <rect {...(draft ?? selectedRect)!} />
                {handles.map((h) => {
                  const r = (draft ?? selectedRect)!;
                  return (
                    <rect
                      key={h}
                      x={r.x + (h.includes("e") ? r.width : 0) - 0.006}
                      y={
                        r.y + (h.includes("s") ? r.height : 0) - 0.006 / aspect
                      }
                      width={0.012}
                      height={0.012 / aspect}
                      data-handle={h}
                      className={`omr-resize-handle omr-handle-${h}`}
                    />
                  );
                })}
              </g>
            )}
          </svg>
        </div>
      </div>
      <p className="muted">
        Select a frame to move it with arrow keys; hold Shift for 10 pixels.
        Escape cancels a drawing. Save calibration to reuse the same layout on
        another sheet.
      </p>
      {selected && selectedRect && (
        <div className="omr-frame-fields">
          <strong>
            {selected.kind === "question"
              ? `Question ${selected.id}`
              : selected.kind === "block"
                ? `Block ${selected.id + 1}`
                : "Answer area"}{" "}
            frame
          </strong>
          {(["x", "y", "width", "height"] as const).map((axis) => (
            <Field key={axis} name={`Selected frame ${axis} percent`}>
              <input
                type="number"
                min={0}
                max={100}
                step={0.01}
                value={Number((selectedRect[axis] * 100).toFixed(3))}
                onChange={(e) => {
                  const value = Number(e.target.value);
                  if (Number.isFinite(value))
                    apply(
                      selected,
                      fitRect({ ...selectedRect, [axis]: value / 100 }),
                    );
                }}
              />
            </Field>
          ))}
          {selected.kind === "question" && (
            <button
              type="button"
              disabled={!c.frames?.some((f) => f.number === selected.id)}
              onClick={() =>
                commit({
                  ...c,
                  frames: c.frames?.filter((f) => f.number !== selected.id),
                })
              }
            >
              Reset question frame
            </button>
          )}
        </div>
      )}
    </section>
  );
}
