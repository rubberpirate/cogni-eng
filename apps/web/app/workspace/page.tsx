import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { Workspace } from "../../components/workspace";
import { apiClient } from "../../../../clients/http";
export const metadata = {
  title: "Workspace",
  robots: { index: false, follow: false },
};
export default async function Page() {
  const session = (await cookies()).get("workspace_session");
  if (!session) redirect("/sign-in");
  const api = apiClient(
    undefined,
    process.env.API_INTERNAL_URL ?? "http://127.0.0.1:8080",
  );
  const options = {
    headers: { Cookie: `workspace_session=${session.value}` },
    cache: "no-store" as const,
    signal: AbortSignal.timeout(8000),
  };
  const [me, contexts] = await Promise.all([
    api.GET("/api/v1/me", options),
    api.GET("/api/v1/workspaces", options),
  ]);
  if (me.response.status === 401 || contexts.response.status === 401)
    redirect("/sign-in");
  if (me.error || contexts.error)
    throw new Error("Workspace service unavailable");
  return <Workspace initialMe={me.data} initialContexts={contexts.data} />;
}
