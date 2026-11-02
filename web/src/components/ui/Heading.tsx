import type { HTMLAttributes } from "react";

type HeadingTag = "h1" | "h2" | "h3" | "h4" | "h5" | "h6";
type HeadingSize = "sm" | "md" | "lg" | "xl";
type HeadingAlign = "left" | "center" | "right";

const sizes: Record<HeadingSize, string> = {
  sm: "text-xl",
  md: "text-2xl",
  lg: "text-3xl",
  xl: "text-5xl",
};

const alignments: Record<HeadingAlign, string> = {
  left: "text-left",
  center: "text-center",
  right: "text-right",
};

type HeadingProps = HTMLAttributes<HTMLHeadingElement> & {
  as?: HeadingTag;
  size?: HeadingSize;
  align?: HeadingAlign;
};

export function Heading({
  as: Tag = "h2",
  size = "md",
  align = "left",
  className = "",
  ...props
}: HeadingProps) {
  return (
    <Tag
      className={`font-semibold text-slate-900 ${sizes[size]} ${alignments[align]} ${className}`}
      {...props}
    />
  );
}
