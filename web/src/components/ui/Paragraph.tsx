import type { HTMLAttributes } from "react";

type ParagraphSize = "sm" | "base" | "lg";

const sizes: Record<ParagraphSize, string> = {
  sm: "text-sm",
  base: "text-base",
  lg: "text-lg",
};

type ParagraphProps = HTMLAttributes<HTMLParagraphElement> & {
  size?: ParagraphSize;
};

export function Paragraph({ size = "base", className = "", ...props }: ParagraphProps) {
  return <p className={`leading-7 text-slate-600 ${sizes[size]} ${className}`} {...props} />;
}
