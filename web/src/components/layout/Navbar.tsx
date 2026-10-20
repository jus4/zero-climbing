import { Link, useNavigate } from "react-router-dom";

import { useAuth } from "../../providers/authContext";
import { Button } from "../ui/Button";

export function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate("/login", { replace: true });
  }

  return (
    <header className="border-b border-slate-200 bg-white">
      <nav className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
        <Link to="/" className="text-lg font-semibold text-slate-900">
          Zero Climbing
        </Link>

        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <>
              <Link
                to="/import"
                className="text-sm font-medium text-slate-700 hover:text-slate-900"
              >
                Import
              </Link>
              <div className="text-right leading-tight">
                <div className="text-sm font-semibold text-slate-800">{user?.alias}</div>
                {user?.email && <div className="text-xs text-slate-400">{user.email}</div>}
              </div>
              <Button variant="secondary" onClick={() => void handleLogout()}>
                Logout
              </Button>
            </>
          ) : (
            <Link to="/login">
              <Button variant="primary">Login</Button>
            </Link>
          )}
        </div>
      </nav>
    </header>
  );
}
