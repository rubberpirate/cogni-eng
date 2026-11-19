// Selection only: Go validates and persists marks and calculates every score.
export function selectQuestionNumbers(
  input: string,
  available: number[],
): number[] {
  const tokens = input.trim().split(",");
  const selected = new Set<number>();
  for (const token of tokens) {
    const match = /^\s*([1-9]\d*)(?:\s*-\s*([1-9]\d*))?\s*$/.exec(token);
    if (!match)
      throw new Error("Use question numbers such as 1-10 or 1,5,8,14,25.");
    const first = Number(match[1]);
    const last = Number(match[2] ?? first);
    if (last < first || !Number.isSafeInteger(last))
      throw new Error("Use an ascending question range, such as 1-10.");
    const matches = available.filter((n) => n >= first && n <= last);
    if (matches.length !== last - first + 1)
      throw new Error(
        `Questions ${token.trim()} are not all available. Check the question numbers.`,
      );
    matches.forEach((n) => selected.add(n));
  }
  return [...selected].sort((a, b) => a - b);
}
