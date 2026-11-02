"use client";

import { Link } from "@/i18n/navigation";
import { ChevronLeftIcon } from "@/icons";
import { cn } from "@/utils";
import SignInFormFields from "./form";
import SignInHeader from "./header";

// Admin-only: tidak ada social login (Google khusus customer mobile)
// dan tidak ada link signup (register publik dimatikan).
export default function SignIn() {
  return (
    <div className={cn("flex w-full flex-1 flex-col lg:w-1/2")}>
      <div className="mx-auto mb-5 w-full max-w-md sm:pt-10">
        <Link
          href="/"
          className="inline-flex items-center text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
        >
          <ChevronLeftIcon className="rtl:rotate-180" />
          Back to home
        </Link>
      </div>

      <div className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center">
        <div>
          <SignInHeader />

          <div>
            <SignInFormFields />
          </div>
        </div>
      </div>
    </div>
  );
}
