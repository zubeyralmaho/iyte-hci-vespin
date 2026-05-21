import { Link } from "expo-router";
import { Pressable, Text, View } from "react-native";
import { ApiError } from "@/api/error";
import { useTranslation } from "@/providers/I18nProvider";
import { useContinueAsGuest } from "../hooks/useAuthActions";

export function GuestAccess() {
  const { t } = useTranslation();
  const guest = useContinueAsGuest();
  const errorMessage = guest.error instanceof ApiError ? guest.error.message : guest.error ? t("auth.errors.generic") : null;

  return (
    <View className="flex-1 justify-center bg-background p-6">
      <View className="gap-6">
        <View className="gap-2">
          <Text className="text-3xl font-semibold text-ink">{t("auth.guest.title")}</Text>
          <Text className="text-base text-muted">{t("auth.guest.subtitle")}</Text>
        </View>

        {errorMessage ? <Text className="text-sm text-danger">{errorMessage}</Text> : null}

        <Pressable
          accessibilityRole="button"
          disabled={guest.isPending}
          onPress={() => guest.mutate()}
          className="items-center rounded-md bg-primary px-4 py-3 disabled:opacity-60"
        >
          <Text className="font-semibold text-onPrimary">
            {guest.isPending ? t("auth.guest.submitting") : t("auth.guest.submit")}
          </Text>
        </Pressable>

        <Link href="/(auth)/chooser" asChild>
          <Pressable accessibilityRole="button">
            <Text className="text-center font-medium text-primary">{t("auth.guest.chooserLink")}</Text>
          </Pressable>
        </Link>
      </View>
    </View>
  );
}
