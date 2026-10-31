"use client";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@/components/ui/input-group";
import { EyeCloseIcon, EyeIcon } from "@/icons";
import { useAuth } from "@/lib/auth-context";
import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useRouter } from "@/i18n/navigation";
import { signUpSchema, type SignUpFormData } from "./schema";

export default function SignUpFormFields() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);
  const { register: signUp } = useAuth();
  const router = useRouter();

  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<SignUpFormData>({
    resolver: zodResolver(signUpSchema),
    defaultValues: {
      fname: "",
      lname: "",
      email: "",
      password: "",
      confirmPassword: "",
      terms: false,
    },
    mode: "onTouched",
  });

  const handleFormSubmit = async (data: SignUpFormData) => {
    setServerError(null);
    try {
      const fullName = `${data.fname} ${data.lname}`.trim();
      // Backend mewajibkan username unik; pakai bagian lokal email.
      const username = data.email.split("@")[0];
      await signUp({
        email: data.email,
        username,
        password: data.password,
        full_name: fullName || undefined,
      });
      router.push("/dashboard");
    } catch (err) {
      setServerError(err instanceof Error ? err.message : "Sign up failed");
    }
  };

  return (
    <form
      onSubmit={handleSubmit(handleFormSubmit)}
      className="space-y-5"
      noValidate
    >
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
        <Field data-invalid={Boolean(errors.fname)}>
          <FieldLabel className="mb-1.5 block">
            First Name <span className="text-red-500">*</span>
          </FieldLabel>
          <Input
            {...register("fname")}
            type="text"
            placeholder="Enter your first name"
            aria-invalid={errors.fname ? "true" : undefined}
          />
          <FieldError errors={[errors.fname]} />
        </Field>

        <Field data-invalid={Boolean(errors.lname)}>
          <FieldLabel className="mb-1.5 block">
            Last Name <span className="text-red-500">*</span>
          </FieldLabel>
          <Input
            {...register("lname")}
            type="text"
            placeholder="Enter your last name"
            aria-invalid={errors.lname ? "true" : undefined}
          />
          <FieldError errors={[errors.lname]} />
        </Field>
      </div>

      <Field data-invalid={Boolean(errors.email)}>
        <FieldLabel className="mb-1.5 block">
          Email <span className="text-red-500">*</span>
        </FieldLabel>
        <Input
          {...register("email")}
          type="email"
          placeholder="Enter your email"
          aria-invalid={errors.email ? "true" : undefined}
        />
        <FieldError errors={[errors.email]} />
      </Field>

      <Field data-invalid={Boolean(errors.password)}>
        <FieldLabel className="mb-1.5 block">
          Password <span className="text-red-500">*</span>
        </FieldLabel>
        <InputGroup>
          <InputGroupInput
            {...register("password")}
            type={showPassword ? "text" : "password"}
            placeholder="Enter your password"
            aria-invalid={errors.password ? "true" : undefined}
          />
          <InputGroupAddon align="inline-end">
            <InputGroupButton
              size="icon-xs"
              onClick={() => setShowPassword((prev) => !prev)}
              aria-label={showPassword ? "Hide password" : "Show password"}
              className="text-gray-500 dark:text-gray-400"
            >
              {showPassword ? (
                <EyeIcon className="fill-current" />
              ) : (
                <EyeCloseIcon className="fill-current" />
              )}
            </InputGroupButton>
          </InputGroupAddon>
        </InputGroup>
        <FieldError errors={[errors.password]} />
      </Field>

      <Field data-invalid={Boolean(errors.confirmPassword)}>
        <FieldLabel className="mb-1.5 block">
          Confirm Password <span className="text-red-500">*</span>
        </FieldLabel>
        <InputGroup>
          <InputGroupInput
            {...register("confirmPassword")}
            type={showConfirmPassword ? "text" : "password"}
            placeholder="Re-enter your password"
            aria-invalid={errors.confirmPassword ? "true" : undefined}
          />
          <InputGroupAddon align="inline-end">
            <InputGroupButton
              size="icon-xs"
              onClick={() => setShowConfirmPassword((prev) => !prev)}
              aria-label={
                showConfirmPassword
                  ? "Hide confirm password"
                  : "Show confirm password"
              }
              className="text-gray-500 dark:text-gray-400"
            >
              {showConfirmPassword ? (
                <EyeIcon className="fill-current" />
              ) : (
                <EyeCloseIcon className="fill-current" />
              )}
            </InputGroupButton>
          </InputGroupAddon>
        </InputGroup>
        <FieldError errors={[errors.confirmPassword]} />
      </Field>

      <Field
        orientation="horizontal"
        className="items-start gap-3"
        data-invalid={Boolean(errors.terms)}
      >
        <Controller
          control={control}
          name="terms"
          render={({ field }) => (
            <Checkbox
              id="terms"
              className="mt-0.5"
              checked={field.value}
              onCheckedChange={(checked) => field.onChange(Boolean(checked))}
              aria-invalid={errors.terms ? "true" : undefined}
            />
          )}
        />
        <div className="flex flex-col gap-1">
          <FieldLabel
            htmlFor="terms"
            className="inline-block cursor-pointer text-sm font-normal text-gray-500 dark:text-gray-400"
          >
            By creating an account means you agree to the{" "}
            <span className="text-gray-800 dark:text-white/90">
              Terms and Conditions,
            </span>{" "}
            and our{" "}
            <span className="text-gray-800 dark:text-white">
              Privacy Policy
            </span>
          </FieldLabel>
          <FieldError errors={[errors.terms]} />
        </div>
      </Field>

      <div>
        {serverError && (
          <p
            role="alert"
            className="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-400"
          >
            {serverError}
          </p>
        )}
        <Button
          className="w-full"
          size="lg"
          type="submit"
          disabled={isSubmitting}
        >
          {isSubmitting ? "Signing Up..." : "Sign Up"}
        </Button>
      </div>
    </form>
  );
}
