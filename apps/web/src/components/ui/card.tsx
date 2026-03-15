import type { HTMLAttributes } from "react";

import { cn } from "@/lib/utils/cn";

/**
 * Card renders a reusable elevated surface.
 */
export function Card({ className, ...props }: HTMLAttributes<HTMLElement>) {
  return <section className={cn("bk-card", className)} {...props} />;
}

/**
 * CardHeader renders the top section of a card.
 */
export function CardHeader({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("grid gap-1.5 p-6 sm:p-8", className)} {...props} />
  );
}

/**
 * CardTitle renders card title text.
 */
export function CardTitle({
  className,
  ...props
}: HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn("bk-heading-card", className)} {...props} />;
}

/**
 * CardDescription renders supporting text below the title.
 */
export function CardDescription({
  className,
  ...props
}: HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("bk-subtle", className)} {...props} />;
}

/**
 * CardContent renders card body content.
 */
export function CardContent({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("px-6 pb-6 sm:px-8 sm:pb-8", className)} {...props} />
  );
}

/**
 * CardFooter renders card bottom actions.
 */
export function CardFooter({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn("px-6 pb-6 pt-2 sm:px-8 sm:pb-8", className)}
      {...props}
    />
  );
}
