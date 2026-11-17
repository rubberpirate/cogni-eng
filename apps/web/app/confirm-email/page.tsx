import { EmailConfirmation } from "../../components/email-confirmation";
export const metadata = {
  title: "Confirm email",
  robots: { index: false, follow: false },
};
export default function Page() {
  return <EmailConfirmation />;
}
