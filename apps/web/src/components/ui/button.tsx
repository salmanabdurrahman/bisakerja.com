import type { ButtonHTMLAttributes, PropsWithChildren } from "react";
import Link from "next/link";

import { cn } from "@/lib/utils/cn";

type ButtonVariant = "primary" | "secondary" | "outline" | "ghost" | "danger";
type ButtonSize = "sm" | "md" | "lg";

interface ButtonStyleOptions {
  variant?: ButtonVariant;
  size?: ButtonSize;
  fullWidth?: boolean;
  className?: string;
}

const variantClass: Record<ButtonVariant, string> = {
  primary:
    "bg-[#000000] text-white hover:bg-[#1A1A1A] border border-transparent focus-visible:ring-black",
  secondary:
    "border border-[#E5E5E5] bg-white text-black hover:bg-[#F9F9F9] focus-visible:ring-black",
  outline:
    "border border-[#E5E5E5] bg-transparent text-black hover:bg-[#F9F9F9] focus-visible:ring-black",
  ghost:
    "bg-transparent text-[#666666] hover:text-black hover:bg-[#F9F9F9] focus-visible:ring-black",
  danger:
    "bg-red-600 text-white hover:bg-red-500 border border-transparent focus-visible:ring-red-600",
};

const sizeClass: Record<ButtonSize, string> = {
  sm: "h-9 rounded-full px-4 text-xs font-normal",
  md: "h-[48px] rounded-full px-6 text-[14px] font-normal",
  lg: "h-[56px] rounded-full px-8 text-base font-normal",
};

const sharedClass =
  "inline-flex items-center justify-center gap-2 transition disabled:pointer-events-none disabled:opacity-60 focus-visible:outline-none focus-visible:ring-1";
/**
 * buttonVariants returns class names for shadcn-like button variants.
 */
export function buttonVariants({
  variant = "primary",
  size = "md",
  fullWidth = false,
  className,
}: ButtonStyleOptions = {}): string {
  return cn(
    sharedClass,
    variantClass[variant],
    sizeClass[size],
    fullWidth ? "w-full" : undefined,
    className,
  );
}

interface ButtonProps
  extends
    Omit<ButtonHTMLAttributes<HTMLButtonElement>, "className">,
    ButtonStyleOptions {}

/**
 * Button renders a style-safe action button.
 */
export function Button({
  variant = "primary",
  size = "md",
  fullWidth = false,
  className,
  type = "button",
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={buttonVariants({ variant, size, fullWidth, className })}
      {...props}
    />
  );
}

interface ButtonLinkProps
  extends
    Omit<React.ComponentProps<typeof Link>, "className">,
    ButtonStyleOptions {}

/**
 * ButtonLink renders a Next.js Link with button styling.
 */
export function ButtonLink({
  variant = "outline",
  size = "md",
  fullWidth = false,
  className,
  ...props
}: PropsWithChildren<ButtonLinkProps>) {
  return (
    <Link
      className={buttonVariants({ variant, size, fullWidth, className })}
      {...props}
    />
  );
}
