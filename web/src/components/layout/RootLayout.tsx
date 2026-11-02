import { Outlet } from "react-router-dom";

import { Navbar } from "./Navbar";

export function RootLayout() {
  return (
    <div className="flex h-dvh flex-col bg-white text-slate-800">
      <Navbar />
      <main className="flex min-h-0 flex-1 flex-col overflow-y-auto">
        <Outlet />
      </main>
    </div>
  );
}
