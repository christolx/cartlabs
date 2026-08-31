import type { Metadata } from "next";
import { PageHeading } from "@/components/marketplace-ui";

export const metadata: Metadata = {
  title: "Docs",
  description: "Buyer, seller, and admin workflows for Cartlabs.",
};

const workflows = [
  {
    id: "buyer",
    role: "Buyer",
    title: "Shop across stores",
    summary:
      "Build one cart, pay once, and follow each seller order separately.",
    steps: [
      ["Browse products", "Find published products from approved stores."],
      [
        "Build a cart",
        "Choose a variant and quantity, then add items to the cart.",
      ],
      ["Review sellers", "Check cart lines grouped by seller before checkout."],
      [
        "Check out",
        "Review cart contents. Inventory becomes reserved when checkout is confirmed.",
      ],
      ["Pay", "Complete mock payment with a success or failure result."],
      [
        "Track the outcome",
        "Follow each seller order, cancel when eligible, or review a delivered item once.",
      ],
    ],
  },
  {
    id: "seller",
    role: "Seller",
    title: "Run one store",
    summary:
      "Manage your catalog and fulfill only orders belonging to your store.",
    steps: [
      [
        "Set up the store",
        "Sign in as seller and create or update the store profile.",
      ],
      [
        "Manage the catalog",
        "Create products, variants, inventory, and product images.",
      ],
      [
        "Publish listings",
        "Publish complete products after store approval. Archive unavailable items.",
      ],
      ["Receive orders", "See only orders belonging to the owned store."],
      [
        "Fulfill orders",
        "Move paid orders through processing, shipped, and delivered.",
      ],
      [
        "Handle cancellations",
        "Cancel paid or processing orders with a reason. Eligible stock is restored.",
      ],
    ],
  },
  {
    id: "admin",
    role: "Admin",
    title: "Keep the marketplace healthy",
    summary:
      "Review platform activity and control stores, listings, users, and audit history.",
    steps: [
      [
        "Monitor the platform",
        "Review platform totals and active gross merchandise value.",
      ],
      [
        "Approve stores",
        "Approve or reject stores before they can publish complete products.",
      ],
      ["Moderate listings", "Suspend or reinstate product listings."],
      [
        "Manage users",
        "Suspend or reactivate users while preserving one active admin.",
      ],
      [
        "Review audit history",
        "Inspect the append-only audit timeline across moderation and fulfillment.",
      ],
    ],
  },
] as const;

export default function DocsPage() {
  return (
    <main className="workspace-page shell info-page docs-page">
      <PageHeading
        family="DOCS / ROLE WORKFLOWS"
        index="02"
        title="Use Cartlabs by role."
        description="Choose a role. Follow its path through the marketplace."
        meta={
          <>
            <span>03 workflows</span>
            <span>Buyer / seller / admin</span>
          </>
        }
      />

      <nav
        className="workspace-subnav info-role-nav"
        aria-label="Role workflows"
      >
        <span className="workspace-subnav-label">
          <span>03</span>
          Roles
        </span>
        <div className="workspace-subnav-links">
          {workflows.map(({ id, role }) => (
            <a key={id} href={`#${id}`}>
              {role}
            </a>
          ))}
        </div>
      </nav>

      <div className="info-stack docs-workflows">
        {workflows.map(({ id, role, title, summary, steps }) => (
          <section
            className="workspace-panel info-panel docs-workflow"
            id={id}
            key={id}
            aria-labelledby={`${id}-title`}
          >
            <div className="panel-heading-row">
              <div>
                <span className="route-family">{role} workflow</span>
                <h2 id={`${id}-title`}>{title}</h2>
                <p>{summary}</p>
              </div>
              <span className="count-badge">
                {String(steps.length).padStart(2, "0")} steps
              </span>
            </div>
            <ol className="record-list info-step-list">
              {steps.map(([stepTitle, body], index) => (
                <li className="record-row info-step-row" key={stepTitle}>
                  <span className="info-step-number" aria-hidden="true">
                    {String(index + 1).padStart(2, "0")}
                  </span>
                  <div>
                    <strong>{stepTitle}</strong>
                    <span>{body}</span>
                  </div>
                </li>
              ))}
            </ol>
          </section>
        ))}
      </div>
    </main>
  );
}
