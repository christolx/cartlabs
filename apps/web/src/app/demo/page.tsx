import { DemoConsole } from "@/components/demo-console";

export const metadata = { title: "Role demo", description: "Exercise buyer, seller, and admin marketplace workflows." };

export default function DemoPage() {
  return <main className="demo-page shell"><div className="demo-intro"><p className="eyebrow">Working role isolation</p><h1>Run marketplace workflow.</h1><p>Create as seller, approve as admin, discover as buyer. Every action reaches same API used by end-to-end tests.</p></div><DemoConsole /></main>;
}
