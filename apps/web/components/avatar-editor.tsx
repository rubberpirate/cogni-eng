"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { apiClient, type components } from "../../../clients/http";
export function AvatarEditor({ csrf }: { csrf: string }) {
  const [ready, setReady] = useState(false),
    [current, setCurrent] = useState(""),
    [source, setSource] = useState(""),
    [zoom, setZoom] = useState(1),
    [x, setX] = useState(50),
    [y, setY] = useState(50),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [busy, setBusy] = useState(false),
    [history, setHistory] = useState<components["schemas"]["ImageHistory"][]>(
      [],
    );
  const canvas = useRef<HTMLCanvasElement>(null);
  const refresh = useCallback(async () => {
    const response = await fetch("/api/v1/profile/avatar", {
      credentials: "include",
      cache: "no-store",
    });
    if (response.ok) {
      const url = URL.createObjectURL(await response.blob());
      setCurrent((previous) => {
        if (previous) URL.revokeObjectURL(previous);
        return url;
      });
    } else if (response.status === 404) {
      setCurrent((previous) => {
        if (previous) URL.revokeObjectURL(previous);
        return "";
      });
    } else {
      const p = await response.json();
      throw new Error(p.message ?? "Image storage unavailable.");
    }
    const h = await apiClient(csrf).GET("/api/v1/profile/image-history");
    if (h.error) throw new Error(h.error.message);
    setHistory(h.data);
  }, [csrf]);
  useEffect(() => {
    void refresh().catch((e) => setError(e.message));
  }, [refresh]);
  useEffect(
    () => () => {
      if (source) URL.revokeObjectURL(source);
    },
    [source],
  );
  useEffect(
    () => () => {
      if (current) URL.revokeObjectURL(current);
    },
    [current],
  );
  useEffect(() => {
    if (!source) return;
    setReady(false);
    let active = true;
    const image = new Image();
    image.onload = () => {
      if (!active) return;
      const ctx = canvas.current?.getContext("2d");
      if (!ctx) return;
      const side = Math.min(image.width, image.height) / zoom;
      ctx.clearRect(0, 0, 256, 256);
      ctx.drawImage(
        image,
        ((image.width - side) * x) / 100,
        ((image.height - side) * y) / 100,
        side,
        side,
        0,
        0,
        256,
        256,
      );
      setReady(true);
    };
    image.onerror = () => setError("The selected image could not be read.");
    image.src = source;
    return () => {
      active = false;
    };
  }, [source, zoom, x, y]);
  async function save() {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const image = await new Promise<Blob | null>((resolve) =>
        canvas.current?.toBlob(resolve, "image/png"),
      );
      if (!image)
        throw new Error("Could not prepare the crop. Choose the image again.");
      const r = await fetch("/api/v1/profile/avatar", {
        method: "PUT",
        credentials: "include",
        headers: { "X-CSRF-Token": csrf, "Content-Type": "image/png" },
        body: image,
      });
      if (!r.ok) {
        const p = await r.json();
        throw new Error(p.message ?? "Image not saved. Retry later.");
      }
      await refresh();
      setSource("");
      setNotice(
        "Profile image saved. The old image is scheduled for private cleanup.",
      );
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Image not saved. Your crop is retained.",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="editor">
      <h2>Private profile image</h2>
      <p>
        Choose a PNG or JPEG, adjust the crop, then save. The original stays on
        your device. The saved image is private to your account.
      </p>
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {notice && (
        <p role="status" className="notice">
          {notice}
        </p>
      )}
      <div className="avatar-work">
        {current ? (
          <img
            className="avatar"
            src={current}
            width={128}
            height={128}
            alt="Your saved profile image"
          />
        ) : (
          <p>No saved profile image.</p>
        )}
        <label>
          Choose profile image
          <input
            type="file"
            accept="image/png,image/jpeg"
            disabled={busy}
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (!file) return;
              setError("");
              if (
                !["image/png", "image/jpeg"].includes(file.type) ||
                file.size > 4 * 1024 * 1024
              ) {
                setError("Choose a PNG or JPEG under 4 MiB.");
                return;
              }
              setReady(false);
              setSource(URL.createObjectURL(file));
              setZoom(1);
              setX(50);
              setY(50);
            }}
          />
        </label>
      </div>
      {source && (
        <div className="crop-editor">
          <canvas
            ref={canvas}
            width={256}
            height={256}
            role="img"
            aria-label="Preview of the square profile crop"
          />
          <div>
            <label>
              Crop zoom
              <input
                type="range"
                min={1}
                max={3}
                step={0.05}
                value={zoom}
                onChange={(e) => {
                  setReady(false);
                  setZoom(Number(e.target.value));
                }}
              />
            </label>
            <label>
              Horizontal crop position
              <input
                type="range"
                min={0}
                max={100}
                value={x}
                onChange={(e) => {
                  setReady(false);
                  setX(Number(e.target.value));
                }}
              />
            </label>
            <label>
              Vertical crop position
              <input
                type="range"
                min={0}
                max={100}
                value={y}
                onChange={(e) => {
                  setReady(false);
                  setY(Number(e.target.value));
                }}
              />
            </label>
            <button
              className="primary"
              disabled={busy || !ready}
              onClick={() => void save()}
            >
              Save cropped image
            </button>
            <button disabled={busy} onClick={() => setSource("")}>
              Discard crop
            </button>
          </div>
        </div>
      )}
      {current && (
        <button
          disabled={busy}
          onClick={() => {
            if (!confirm("Remove your saved profile image?")) return;
            setBusy(true);
            void apiClient(csrf)
              .DELETE("/api/v1/profile/avatar", {
                params: { header: { "X-CSRF-Token": csrf } },
              })
              .then(async (r) => {
                if (r.error) throw new Error(r.error.message);
                await refresh();
                setNotice(
                  "Image removed from your profile; private object cleanup is queued.",
                );
              })
              .catch((e) => setError(e.message))
              .finally(() => setBusy(false));
          }}
        >
          Remove saved image
        </button>
      )}
      <details>
        <summary>Image storage history</summary>
        <p>
          Replacement and abandoned uploads are purged by the worker. A failed
          cleanup can be retried here.
        </p>
        <ul className="activity-list">
          {history.map((h) => (
            <li key={h.id}>
              <strong>{h.state}</strong> · {h.bytes} bytes ·{" "}
              {new Date(h.created_at).toLocaleString()}
              {h.error && <p>{h.error}</p>}
              {["staged", "retired"].includes(h.state) && (
                <button
                  onClick={() =>
                    void apiClient(csrf)
                      .POST("/api/v1/profile/image-history/{id}/retry", {
                        params: {
                          path: { id: h.id },
                          header: { "X-CSRF-Token": csrf },
                        },
                      })
                      .then(async (r) => {
                        if (r.error) throw new Error(r.error.message);
                        await refresh();
                        setNotice("Cleanup scheduled.");
                      })
                      .catch((e) => setError(e.message))
                  }
                >
                  Retry image cleanup
                </button>
              )}
            </li>
          ))}
        </ul>
      </details>
    </section>
  );
}
