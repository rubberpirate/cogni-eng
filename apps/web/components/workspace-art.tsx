import type { CSSProperties } from "react";

// Original line artwork; no institution or learner records are embedded.
export function StudyArtwork({ compact = false }: { compact?: boolean }) {
  return (
    <svg
      className={compact ? "study-art compact" : "study-art"}
      viewBox="0 0 400 260"
      fill="none"
      aria-hidden="true"
    >
      <circle cx="219" cy="133" r="102" fill="currentColor" opacity=".07" />
      <path
        d="M31 223H365M63 200V223M334 199V223"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <g transform="rotate(-12 142 130)">
        <rect
          x="73"
          y="55"
          width="130"
          height="158"
          rx="12"
          fill="var(--surface)"
          stroke="currentColor"
          strokeWidth="2"
        />
        <path
          d="M94 82H158M94 95H178M94 109H141"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
        />
        {[132, 158, 184].map((y) => (
          <g key={y}>
            <path d={`M94 ${y}H116`} stroke="currentColor" strokeWidth="2" />
            {[136, 156, 176].map((x) => (
              <circle
                key={x}
                cx={x}
                cy={y}
                r="5"
                stroke="currentColor"
                strokeWidth="1.5"
                fill={x === 156 ? "currentColor" : "none"}
              />
            ))}
          </g>
        ))}
      </g>
      <g transform="rotate(9 253 146)">
        <rect
          x="202"
          y="75"
          width="112"
          height="143"
          rx="12"
          fill="var(--surface)"
          stroke="currentColor"
          strokeWidth="2"
        />
        <rect
          x="219"
          y="94"
          width="40"
          height="7"
          rx="3.5"
          fill="currentColor"
          opacity=".2"
        />
        {[122, 153, 184].map((y) => (
          <g key={y}>
            <rect
              x="220"
              y={y - 7}
              width="14"
              height="14"
              rx="4"
              stroke="currentColor"
              strokeWidth="1.5"
            />
            <path
              d={`M224 ${y}l3 3 5-6M245 ${y - 2}H294M245 ${y + 5}H278`}
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </g>
        ))}
      </g>
      <circle cx="298" cy="58" r="22" fill="var(--highlight)" />
      <path
        d="M288 58l7 7 14-16M51 83v12M45 89h12M338 132v12M332 138h12"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M165 36l5-10 5 10-5 10zM340 184l5-10 5 10-5 10z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
    </svg>
  );
}
const paths: Record<string, string> = {
  home: "M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z",
  generate: "m12 3 2.5 6.5L21 12l-6.5 2.5L12 21l-2.5-6.5L3 12l6.5-2.5z",
  curriculum:
    "M4 4h6a3 3 0 0 1 3 3v14a4 4 0 0 0-4-3H4zM13 7a3 3 0 0 1 3-3h5v14h-4a4 4 0 0 0-4 3",
  papers: "M6 3h9l4 4v14H6zM14 3v5h5M9 12h7M9 16h5",
  scan: "M8 3H3v5M16 3h5v5M3 16v5h5M21 16v5h-5M3 12h18M8 8h8v8H8z",
  results: "M4 20V10h4v10M10 20V4h4v16M16 20v-7h4v7",
  samples:
    "M3 8h18v13H3zM2 5h20v4H2zM12 5v16M12 5C5 5 6 0 9 2l3 3c7 0 6-5 3-3z",
  jobs: "M20 7v5h-5M4 17v-5h5M5 8a8 8 0 0 1 13-3l2 3M4 16l2 3a8 8 0 0 0 13-3",
  account: "M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0M4 21v-3a8 8 0 0 1 16 0v3",
  security: "m12 2 8 4v6c0 5-8 10-8 10S4 17 4 12V6zM8 12l3 3 5-6",
  arrow: "M5 12h14M13 6l6 6-6 6",
};
export function WorkIcon({
  name,
  style,
}: {
  name: string;
  style?: CSSProperties;
}) {
  return (
    <svg
      className="work-icon"
      style={style}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={paths[name] ?? paths.papers} />
    </svg>
  );
}
