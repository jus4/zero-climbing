import { createBrowserRouter } from "react-router-dom";

import { OnboardingRoute, ProtectedRoute, PublicOnlyRoute } from "./components/guards";
import { RootLayout } from "./components/layout/RootLayout";
import { ImportPage } from "./pages/ImportPage";
import { LoginPage } from "./pages/LoginPage";
import { MapPage } from "./pages/MapPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { OnboardingPage } from "./pages/OnboardingPage";

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      {
        path: "/login",
        element: (
          <PublicOnlyRoute>
            <LoginPage />
          </PublicOnlyRoute>
        ),
      },
      {
        path: "/onboarding",
        element: (
          <OnboardingRoute>
            <OnboardingPage />
          </OnboardingRoute>
        ),
      },
      {
        path: "/",
        element: (
          <ProtectedRoute>
            <MapPage />
          </ProtectedRoute>
        ),
      },
      {
        path: "/import",
        element: (
          <ProtectedRoute>
            <ImportPage />
          </ProtectedRoute>
        ),
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
