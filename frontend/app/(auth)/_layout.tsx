import { Redirect, Stack } from "expo-router";
import { useAuthStore } from "@/features/auth/store";

export default function AuthLayout() {
  const token = useAuthStore((s) => s.token);
  const isHydrated = useAuthStore((s) => s.isHydrated);

  if (!isHydrated) return null;
  if (token) return <Redirect href="/(app)/(tabs)/devices" />;
  return (
    <Stack
      screenOptions={{
        headerShown: false,
        contentStyle: { backgroundColor: "#460812" },
      }}
    />
  );
}
