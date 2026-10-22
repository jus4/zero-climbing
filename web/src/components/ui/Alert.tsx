import type { HTMLAttributes } from "react";

type Tone = "error" | "success" | "info";

const tones: Record<Tone, string> = {
  error: "border-red-200 bg-red-50 text-red-700",
  success: "border-green-200 bg-green-50 text-green-700",
  info: "border-slate-200 bg-slate-50 text-slate-700",
};

type AlertProps = HTMLAttributes<HTMLDivElement> & {
  tone?: Tone;
};

export function Alert({ tone = "info", className = "", ...props }: AlertProps) {
  return (
    <div
      role="alert"
      className={`rounded-md border px-3 py-2 text-sm ${tones[tone]} ${className}`}
      {...props}
    />
  );
}
