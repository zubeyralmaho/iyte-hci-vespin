import { Pressable, Text, View } from "react-native";
import { useTranslation } from "@/providers/I18nProvider";
import { useLogout } from "../hooks/useAuthActions";
import { useAuthStore } from "../store";

export function AccountPanel() {
  const { t } = useTranslation();
  const currentUser = useAuthStore((s) => s.currentUser);
  const logout = useLogout();

  return (
    <View className="flex-1 justify-center bg-background p-6">
      <View className="gap-4">
        <Text className="text-3xl font-semibold text-ink">{t("settings.title")}</Text>
        <View className="gap-1">
          <Text className="text-base text-ink">{currentUser?.displayName ?? currentUser?.email ?? t("auth.account.guest")}</Text>
          <Text className="text-sm text-muted">{t(`auth.roles.${currentUser?.role ?? "guest"}`)}</Text>
        </View>
        <Pressable
          accessibilityRole="button"
          disabled={logout.isPending}
          onPress={() => logout.mutate()}
          className="items-center rounded-md border border-border px-4 py-3 disabled:opacity-60"
        >
          <Text className="font-semibold text-ink">
            {logout.isPending ? t("auth.logout.submitting") : t("auth.logout.submit")}
          </Text>
        </Pressable>
      </View>
    </View>
  );
}
