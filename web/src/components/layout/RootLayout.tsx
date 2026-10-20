import { Outlet } from "react-router-dom";

import { Navbar } from "./Navbar";

export function RootLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50 text-slate-800">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  );
}
