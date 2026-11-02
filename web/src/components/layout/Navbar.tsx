import { Link, NavLink, useNavigate } from "react-router-dom";

import { useAuth } from "../../providers/authContext";
import { Button } from "../ui/Button";

const navLinkBase =
  "inline-flex items-center rounded-md border px-3 py-2 text-lg font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2";
const navLinkInactive =
  "border-transparent text-slate-600 hover:border-primary-100 hover:bg-primary-50";
const navLinkActive = "border-primary-100 bg-primary-50 text-slate-600";

export function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate("/login", { replace: true });
  }

  return (
    <header className="border-b border-slate-200 bg-white">
      <nav>
        {isAuthenticated && (
          <div className="border-b border-primary-100 bg-primary-50/70">
            <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-end gap-4 px-4 py-2">
              <div className="text-right leading-tight">
                <div className="text-sm font-semibold text-slate-800">{user?.alias}</div>
                {user?.email && <div className="text-xs text-slate-500">{user.email}</div>}
              </div>
              <Button variant="secondary" onClick={() => void handleLogout()}>
                Kirjaudu ulos
              </Button>
            </div>
          </div>
        )}
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-x-4 gap-y-2 px-4 py-3">
          <Link
            to="/"
            className="font-heading text-lg font-bold tracking-tight text-primary-700 transition-colors hover:text-primary-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2"
          >
            Zero Climbing
          </Link>

          <div className="flex flex-wrap items-center gap-4">
            <NavLink
              to="/about"
              end
              className={({ isActive }) =>
                `${navLinkBase} ${isActive ? navLinkActive : navLinkInactive}`
              }
            >
              Miksi?
            </NavLink>

            {isAuthenticated && (
              <>
                <NavLink
                  to="/import"
                  end
                  className={({ isActive }) =>
                    `${navLinkBase} ${isActive ? navLinkActive : navLinkInactive}`
                  }
                >
                  Tuo lokaatiot
                </NavLink>
                <Link to="/import" className={`${navLinkBase} ${navLinkInactive}`}>
                  Hallinnoi
                </Link>
              </>
            )}

            {!isAuthenticated && (
              <Link
                to="/login"
                className={`${navLinkBase} border-primary-600 bg-primary-600 text-white hover:bg-primary-700`}
              >
                Kirjaudu
              </Link>
            )}
          </div>
        </div>
      </nav>
    </header>
  );
}
