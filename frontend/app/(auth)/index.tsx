import { Image, Pressable, View } from "react-native";
import { router } from "expo-router";
import { Screen } from "@/components/layout/Screen";
import { AppText } from "@/components/ui/AppText";
import { Button } from "@/components/ui/Button";
import { useTranslation } from "@/providers/I18nProvider";

const HERO = require("../../assets/images/hero-welcome.png");

export default function WelcomeScreen() {
  const { t } = useTranslation();

  return (
    <Screen tone="primary" padded={false} edges={["top"]}>
      <View className="flex-1">
        <Image
          source={HERO}
          resizeMode="cover"
          className="absolute inset-x-0 bottom-0 top-24 w-full"
          style={{ opacity: 0.95 }}
        />

        <View className="items-center pt-16">
          <AppText variant="display">{t("auth.welcome.title")}</AppText>
        </View>

        <View className="mt-auto items-center gap-2 px-6 pb-6">
          <Pressable accessibilityRole="link" hitSlop={8}>
            <AppText variant="body">{t("auth.welcome.licence")}</AppText>
          </Pressable>
          <Pressable accessibilityRole="link" hitSlop={8}>
            <AppText variant="body">{t("auth.welcome.privacy")}</AppText>
          </Pressable>
          <Button
            label={t("auth.welcome.start")}
            variant="secondary"
            size="lg"
            fullWidth
            className="mt-4"
            onPress={() => router.push("/(auth)/chooser")}
          />
        </View>
      </View>
    </Screen>
  );
}
