"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api, type User } from "@/lib/api";
import { cn } from "@/utils";
import { useEffect, useState } from "react";
import UsersTableRow from "./users-table-row";

export default function UsersTable({ className }: { className?: string }) {
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api
      .listUsers()
      .then((res) => setUsers(res.users))
      .catch((err: unknown) =>
        setError(err instanceof Error ? err.message : "Failed to load users"),
      )
      .finally(() => setIsLoading(false));
  }, []);

  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-white/5 dark:bg-white/3",
        className,
      )}
    >
      <div className="max-w-full overflow-x-auto">
        <Table>
          <TableHeader className="border-b border-gray-100 dark:border-white/5">
            <TableRow>
              <TableCell
                isHeader
                className="px-5 py-3 text-start text-xs font-medium text-gray-500 sm:px-6 dark:text-gray-400"
              >
                User
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 text-start text-xs font-medium text-gray-500 sm:px-6 dark:text-gray-400"
              >
                Email
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 text-start text-xs font-medium text-gray-500 sm:px-6 dark:text-gray-400"
              >
                Role
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 text-start text-xs font-medium text-gray-500 sm:px-6 dark:text-gray-400"
              >
                Joined
              </TableCell>
            </TableRow>
          </TableHeader>

          <TableBody className="divide-y divide-gray-100 dark:divide-white/5">
            {isLoading && (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="px-5 py-8 text-center text-sm text-gray-500 dark:text-gray-400"
                >
                  Loading users...
                </TableCell>
              </TableRow>
            )}
            {!isLoading && error && (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="px-5 py-8 text-center text-sm text-red-600 dark:text-red-400"
                >
                  {error}
                </TableCell>
              </TableRow>
            )}
            {!isLoading && !error && users.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="px-5 py-8 text-center text-sm text-gray-500 dark:text-gray-400"
                >
                  No users found.
                </TableCell>
              </TableRow>
            )}
            {!isLoading &&
              !error &&
              users.map((user) => <UsersTableRow key={user.id} user={user} />)}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
