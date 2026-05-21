import { useEffect, type ReactNode } from "react";
import { configureApiClient } from "@/api/client";
import { useGetUsersMe } from "@/api/generated/users/users";
import { useAuthStore } from "@/features/auth/store";

configureApiClient({
  getAuthToken: () => useAuthStore.getState().token,
  onUnauthorized: () => {
    void useAuthStore.getState().clearSession();
  },
});

type Props = {
  children: ReactNode;
};

export function AuthSessionProvider({ children }: Props) {
  const token = useAuthStore((s) => s.token);
  const isHydrated = useAuthStore((s) => s.isHydrated);
  const hydrateSession = useAuthStore((s) => s.hydrateSession);
  const setCurrentUser = useAuthStore((s) => s.setCurrentUser);

  const meQuery = useGetUsersMe({
    query: {
      enabled: isHydrated && !!token,
      staleTime: 60_000,
    },
  });

  useEffect(() => {
    if (!isHydrated) {
      void hydrateSession();
    }
  }, [hydrateSession, isHydrated]);

  useEffect(() => {
    if (meQuery.data?.status === 200) {
      setCurrentUser(meQuery.data.data);
    }
  }, [meQuery.data, setCurrentUser]);

  return children;
}
