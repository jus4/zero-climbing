import type { HTMLAttributes } from "react";

export function Divider({ className = "", ...props }: HTMLAttributes<HTMLHRElement>) {
  return <hr className={`my-8 border-slate-200 my-4 ${className}`} {...props} />;
}
