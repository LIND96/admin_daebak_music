"use client";

import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { TableCell, TableRow } from "@/components/ui/table";
import type { User } from "@/lib/api";

function getInitials(name: string) {
  const initials = name
    .split(" ")
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase())
    .join("")
    .slice(0, 2);
  return initials || "?";
}

export default function UsersTableRow({ user }: { user: User }) {
  const displayName = user.full_name || user.username;

  return (
    <TableRow className="transition hover:bg-gray-50/50 dark:hover:bg-white/2">
      <TableCell className="px-5 py-4 text-start sm:px-6">
        <div className="flex items-center gap-3">
          <Avatar size="default">
            <AvatarImage src={user.avatar_url} alt={user.username} />
            <AvatarFallback>{getInitials(displayName)}</AvatarFallback>
          </Avatar>
          <div>
            <span className="block text-sm font-medium text-gray-800 dark:text-white/90">
              {displayName}
            </span>
            <span className="block text-xs text-gray-500 dark:text-gray-400">
              @{user.username}
            </span>
          </div>
        </div>
      </TableCell>
      <TableCell className="px-5 py-4 text-start text-sm text-gray-500 sm:px-6 dark:text-gray-400">
        {user.email}
      </TableCell>
      <TableCell className="px-5 py-4 text-start sm:px-6">
        <Badge size="sm" color={user.role === "admin" ? "primary" : "gray"}>
          {user.role}
        </Badge>
      </TableCell>
      <TableCell className="px-5 py-4 text-start text-sm text-gray-500 sm:px-6 dark:text-gray-400">
        {new Date(user.created_at).toLocaleDateString()}
      </TableCell>
    </TableRow>
  );
}
