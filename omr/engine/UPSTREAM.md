# Bundled OMRChecker runtime

Source: https://github.com/Udayraj123/OMRChecker

Revision: `5cf44a5a9c7a49e2e541c6e19aefd09a41a4d494`

License: MIT; see [LICENSE](LICENSE) and [Contributors.md](Contributors.md).

The `src/` runtime files are copied without changes from this revision. Upstream
tests, sample documents, Git metadata, CI and development configuration are
excluded. The application uses its own original synthetic test sheets. `REVISION`
records the upstream version for recognition responses without invoking Git.

Dependencies are installed from the application's frozen `../uv.lock`. The
application adapter in `../bridge.py` handles input validation and headless
recognition; authoritative grading remains in Go and does not call upstream
evaluation code. Keep the upstream copyright and license when distributing this
folder. Changes to bundled runtime source must update this notice and record
their difference from the pinned revision.
