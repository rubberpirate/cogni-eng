import { test, expect } from "bun:test";
import {
  buildGrid,
  gridRows,
  setBlockRect,
  fitRect,
  type Calibration,
} from "../clients/omr-calibration";
const empty: Calibration = { options: 4, radius: 0.008, blocks: [] };
test("visual grids preserve numbering and frames override only the selected answer", () => {
  const c = buildGrid(
    empty,
    { x: 0.1, y: 0.1, width: 0.8, height: 0.8 },
    5,
    2,
    10,
    1.25,
  );
  expect(c.blocks.map((b) => [b.first, b.count])).toEqual([
    [10, 3],
    [13, 2],
  ]);
  expect(gridRows(c, 1.25).map((r) => r.number)).toEqual([10, 11, 12, 13, 14]);
  const changed = {
    ...c,
    frames: [
      { number: 11, rect: { x: 0.2, y: 0.3, width: 0.3, height: 0.02 } },
    ],
  };
  const before = gridRows(c, 1.25),
    after = gridRows(changed, 1.25);
  expect(after[0]).toEqual(before[0]);
  expect(after[2]).toEqual(before[2]);
  expect(after[1].centers[0].x).toBeCloseTo(0.208);
  expect(after[1].centers.at(-1)!.x).toBeCloseTo(0.492);
  expect(after[1].centers.every((p) => p.y === 0.31)).toBe(true);
  expect(
    setBlockRect(changed, 0, { x: 0.1, y: 0.1, width: 0.3, height: 0.7 }, 1.25)
      .frames,
  ).toEqual([]);
});
test("one-question frame centers and moved rectangles stay finite and inside the image", () => {
  const c = buildGrid(
    empty,
    { x: 0.1, y: 0.2, width: 0.8, height: 0.1 },
    1,
    1,
    1,
    1.25,
  );
  const row = gridRows(c, 1.25)[0];
  expect(row.centers[0].y).toBeCloseTo(0.25);
  expect(
    row.centers.every((p) => Number.isFinite(p.x) && Number.isFinite(p.y)),
  ).toBe(true);
  expect(fitRect({ x: 0.95, y: -0.2, width: 0.2, height: 0.3 })).toEqual({
    x: 0.8,
    y: 0,
    width: 0.2,
    height: 0.3,
  });
});
