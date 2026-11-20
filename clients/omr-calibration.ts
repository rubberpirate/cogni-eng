import type { A } from "./academic";
export type Calibration = A<"AcademicGridCalibration">;
export type Rect = A<"AcademicGridRect">;
export type Point = A<"AcademicPoint">;
export const clamp = (n: number, min = 0, max = 1) =>
  Math.max(min, Math.min(max, n));

export function gridRows(c: Calibration, aspect: number) {
  const frames = new Map(c.frames?.map((f) => [f.number, f.rect]));
  return c.blocks.flatMap((block) =>
    Array.from({ length: block.count }, (_, index) => {
      const number = block.first + index;
      const v = index / Math.max(1, block.count - 1);
      const frame = frames.get(number);
      const centers = Array.from({ length: c.options }, (_, option) => {
        const u = option / Math.max(1, c.options - 1);
        return frame
          ? {
              x: frame.x + c.radius + u * (frame.width - 2 * c.radius),
              y: frame.y + frame.height / 2,
            }
          : {
              x:
                (1 - v) * ((1 - u) * block.top_left.x + u * block.top_right.x) +
                v * ((1 - u) * block.bottom_left.x + u * block.bottom_right.x),
              y:
                (1 - v) * ((1 - u) * block.top_left.y + u * block.top_right.y) +
                v * ((1 - u) * block.bottom_left.y + u * block.bottom_right.y),
            };
      });
      const x = Math.min(...centers.map((p) => p.x)) - c.radius;
      const y = Math.min(...centers.map((p) => p.y)) - c.radius / aspect;
      return {
        number,
        centers,
        rect: frame ?? {
          x,
          y,
          width: Math.max(...centers.map((p) => p.x)) + c.radius - x,
          height: Math.max(...centers.map((p) => p.y)) + c.radius / aspect - y,
        },
      };
    }),
  );
}

export function blockRect(c: Calibration, index: number, aspect: number): Rect {
  const b = c.blocks[index];
  const points = [b.top_left, b.top_right, b.bottom_left, b.bottom_right];
  const x = Math.min(...points.map((p) => p.x)) - c.radius;
  const y = Math.min(...points.map((p) => p.y)) - c.radius / aspect;
  return {
    x,
    y,
    width: Math.max(...points.map((p) => p.x)) + c.radius - x,
    height: Math.max(...points.map((p) => p.y)) + c.radius / aspect - y,
  };
}

export function setBlockRect(
  c: Calibration,
  index: number,
  rect: Rect,
  aspect: number,
): Calibration {
  const block = c.blocks[index];
  const ry = c.radius / aspect;
  const top = block.count === 1 ? rect.y + rect.height / 2 : rect.y + ry;
  const bottom = block.count === 1 ? top : rect.y + rect.height - ry;
  return {
    ...c,
    blocks: c.blocks.map((b, i) =>
      i === index
        ? {
            ...b,
            top_left: { x: rect.x + c.radius, y: top },
            top_right: { x: rect.x + rect.width - c.radius, y: top },
            bottom_left: { x: rect.x + c.radius, y: bottom },
            bottom_right: {
              x: rect.x + rect.width - c.radius,
              y: bottom,
            },
          }
        : b,
    ),
    frames: c.frames?.filter(
      (f) => f.number < block.first || f.number >= block.first + block.count,
    ),
  };
}

export function buildGrid(
  c: Calibration,
  area: Rect,
  total: number,
  columns: number,
  first: number,
  aspect: number,
): Calibration {
  let next: Calibration = { ...c, area, blocks: [], frames: [] };
  let number = first;
  for (let i = 0; i < columns; i++) {
    const count = Math.floor(total / columns) + (i < total % columns ? 1 : 0);
    if (!count) continue;
    next.blocks.push({
      first: number,
      count,
      top_left: { x: 0, y: 0 },
      top_right: { x: 1, y: 0 },
      bottom_left: { x: 0, y: 1 },
      bottom_right: { x: 1, y: 1 },
    });
    const gap = columns > 1 ? (area.width / columns) * 0.16 : 0;
    next = setBlockRect(
      next,
      i,
      {
        x: area.x + (i * area.width) / columns + gap / 2,
        y: area.y,
        width: area.width / columns - gap,
        height: area.height,
      },
      aspect,
    );
    number += count;
  }
  return next;
}

export function fitRect(r: Rect): Rect {
  const width = clamp(r.width, 0.001, 1),
    height = clamp(r.height, 0.001, 1);
  return {
    x: clamp(r.x, 0, 1 - width),
    y: clamp(r.y, 0, 1 - height),
    width,
    height,
  };
}
