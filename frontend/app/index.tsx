import { Redirect } from "expo-router";
import { useAuthStore } from "@/features/auth/store";

export default function Index() {
  const token = useAuthStore((s) => s.token);
  const isHydrated = useAuthStore((s) => s.isHydrated);

  if (!isHydrated) return null;
  if (token) return <Redirect href="/(app)/(tabs)/devices" />;
  return <Redirect href="/(auth)" />;
}
