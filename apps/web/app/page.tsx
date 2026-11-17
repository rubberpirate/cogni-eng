import Link from "next/link";
export default function Home() {
  return (
    <main id="main" className="public">
      <header>
        <Link href="/" className="wordmark">
          LW / Learning Workspace
        </Link>
        <Link className="button" href="/sign-in">
          Sign in
        </Link>
      </header>
      <section className="introduction">
        <p className="eyebrow">For institutions, learners and families</p>
        <h1>
          A clear place
          <br />
          to move learning forward.
        </h1>
        <p className="lead">
          Organize your institution, bring people into the right workspace, and
          keep a record of the decisions that matter.
        </p>
        <Link className="button primary" href="/workspace">
          Open your workspace <span aria-hidden="true">↗</span>
        </Link>
      </section>
      <section className="public-note">
        <h2>Built around responsibility</h2>
        <p>
          Access follows your institution, campus and responsibilities. Your
          account can belong to more than one institution without mixing their
          records.
        </p>
        <h2>Development availability</h2>
        <p>
          This installation is under active implementation. Organization and
          access tools are being qualified. Assessment, learning, mobile and
          business workflows are tracked in the product register and are not yet
          available.
        </p>
      </section>
      <footer>
        <Link href="/help">Help and setup</Link>
        <Link href="/policies">Terms and privacy</Link>
      </footer>
    </main>
  );
}
