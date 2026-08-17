import { DemoEntry } from "@/components/demo-entry";

export const metadata = { title: "Demo guide", description: "Enter buyer, seller, and admin marketplace workflows through normal product pages." };

export default function DemoPage() {
  return <main className="demo-page shell"><div className="demo-intro"><p className="eyebrow">Marketplace workflow</p><h1>Run real actor pages.</h1><p>Use roles in order to exercise store verification, publishing, listing enforcement, purchase, fulfillment, and review.</p></div><DemoEntry /></main>;
}
