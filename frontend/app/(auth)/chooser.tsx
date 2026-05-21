import { Image, Pressable, View } from "react-native";
import { router } from "expo-router";
import { Screen } from "@/components/layout/Screen";
import { AppText } from "@/components/ui/AppText";
import { Button } from "@/components/ui/Button";
import { IconButton } from "@/components/ui/IconButton";
import { Mark } from "@/components/brand/Mark";
import { useTranslation } from "@/providers/I18nProvider";

const HERO = require("../../assets/images/hero-auth.png");

export default function ChooserScreen() {
  const { t } = useTranslation();

  return (
    <Screen tone="primary" padded={false} edges={["top", "bottom"]}>
      <View className="flex-1 px-6">
        <View className="items-center pt-2">
          <Mark size="sm" />
        </View>

        <View className="flex-1 items-center justify-center">
          <Image
            source={HERO}
            resizeMode="contain"
            className="h-full w-full"
            style={{ opacity: 0.95 }}
          />
        </View>

        <View className="gap-3 pb-2">
          <Pressable
            accessibilityRole="link"
            hitSlop={8}
            onPress={() => router.push("/(auth)/guest")}
            className="items-center py-2"
          >
            <AppText variant="body">{t("auth.chooser.guest")}</AppText>
          </Pressable>
          <Button
            label={t("auth.chooser.login")}
            variant="secondary"
            size="lg"
            fullWidth
            onPress={() => router.push("/(auth)/login")}
          />
          <Button
            label={t("auth.chooser.register")}
            variant="ghost"
            size="lg"
            fullWidth
            labelTone="onPrimary"
            className="border border-onPrimary/40"
            onPress={() => router.push("/(auth)/register")}
          />

          <AppText variant="caption" className="mt-2 text-center">
            {t("auth.chooser.or")}
          </AppText>

          <View className="flex-row items-center justify-center gap-4">
            <IconButton name="logo-google" accessibilityLabel="Google" tone="onPrimary" />
            <IconButton name="logo-apple" accessibilityLabel="Apple" tone="onPrimary" />
            <IconButton name="logo-facebook" accessibilityLabel="Facebook" tone="onPrimary" />
          </View>
        </View>
      </View>
    </Screen>
  );
}
