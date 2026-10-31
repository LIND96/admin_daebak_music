import ComponentCard from "@/components/common/component-card";
import PageBreadcrumb from "@/components/common/page-breadcrumb";
import UsersTable from "@/components/main/tables/users-table";

import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Users | TailAdmin - Next.js Admin Dashboard Template",
  description: "Manage registered users of the admin panel.",
};

export default function UsersPage() {
  return (
    <div>
      <PageBreadcrumb pageTitle="Users" />
      <div className="space-y-6">
        <ComponentCard title="All Users">
          <UsersTable />
        </ComponentCard>
      </div>
    </div>
  );
}
