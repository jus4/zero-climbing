import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";

import { useAuth } from "../providers/authContext";
import { FullPageSpinner } from "./ui/Spinner";

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, onboardingRequired, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return <FullPageSpinner />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (onboardingRequired && location.pathname !== "/onboarding") {
    return <Navigate to="/onboarding" replace />;
  }

  return <>{children}</>;
}

export function PublicOnlyRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, onboardingRequired, isLoading } = useAuth();

  if (isLoading) {
    return <FullPageSpinner />;
  }

  if (onboardingRequired) {
    return <Navigate to="/onboarding" replace />;
  }

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}

export function OnboardingRoute({ children }: { children: ReactNode }) {
  const { onboardingRequired, isLoading } = useAuth();

  if (isLoading) {
    return <FullPageSpinner />;
  }

  if (!onboardingRequired) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
