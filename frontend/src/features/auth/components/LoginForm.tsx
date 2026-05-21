import { zodResolver } from "@hookform/resolvers/zod";
import { Link } from "expo-router";
import { Controller, useForm } from "react-hook-form";
import { Pressable, Text, TextInput, View } from "react-native";
import { ApiError } from "@/api/error";
import { useTranslation } from "@/providers/I18nProvider";
import { useLogin } from "../hooks/useAuthActions";
import { loginSchema, type LoginInput } from "../schemas/login";

export function LoginForm() {
  const { t } = useTranslation();
  const login = useLogin();
  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const submit = form.handleSubmit((values) => login.mutate(values));
  const errorMessage = login.error instanceof ApiError ? login.error.message : login.error ? t("auth.errors.generic") : null;

  return (
    <View className="flex-1 justify-center bg-background p-6">
      <View className="gap-6">
        <View className="gap-2">
          <Text className="text-3xl font-semibold text-ink">{t("auth.login.title")}</Text>
          <Text className="text-base text-muted">{t("auth.login.subtitle")}</Text>
        </View>

        <View className="gap-4">
          <View className="gap-2">
            <Text className="text-sm font-medium text-ink">{t("auth.fields.email")}</Text>
            <Controller
              control={form.control}
              name="email"
              render={({ field, fieldState }) => (
                <>
                  <TextInput
                    autoCapitalize="none"
                    autoComplete="email"
                    keyboardType="email-address"
                    onBlur={field.onBlur}
                    onChangeText={field.onChange}
                    value={field.value}
                    className="rounded-md border border-border bg-surface px-4 py-3 text-ink"
                  />
                  {fieldState.error?.message ? (
                    <Text className="text-sm text-danger">{t(fieldState.error.message)}</Text>
                  ) : null}
                </>
              )}
            />
          </View>

          <View className="gap-2">
            <Text className="text-sm font-medium text-ink">{t("auth.fields.password")}</Text>
            <Controller
              control={form.control}
              name="password"
              render={({ field, fieldState }) => (
                <>
                  <TextInput
                    autoCapitalize="none"
                    autoComplete="password"
                    onBlur={field.onBlur}
                    onChangeText={field.onChange}
                    secureTextEntry
                    value={field.value}
                    className="rounded-md border border-border bg-surface px-4 py-3 text-ink"
                  />
                  {fieldState.error?.message ? (
                    <Text className="text-sm text-danger">{t(fieldState.error.message)}</Text>
                  ) : null}
                </>
              )}
            />
          </View>

          {errorMessage ? <Text className="text-sm text-danger">{errorMessage}</Text> : null}

          <Pressable
            accessibilityRole="button"
            disabled={login.isPending}
            onPress={submit}
            className="items-center rounded-md bg-primary px-4 py-3 disabled:opacity-60"
          >
            <Text className="font-semibold text-onPrimary">
              {login.isPending ? t("auth.login.submitting") : t("auth.login.submit")}
            </Text>
          </Pressable>
        </View>

        <View className="gap-3">
          <Link href="/(auth)/register" asChild>
            <Pressable accessibilityRole="button">
              <Text className="text-center font-medium text-primary">{t("auth.login.registerLink")}</Text>
            </Pressable>
          </Link>
          <Link href="/(auth)/guest" asChild>
            <Pressable accessibilityRole="button">
              <Text className="text-center text-muted">{t("auth.login.guestLink")}</Text>
            </Pressable>
          </Link>
        </View>
      </View>
    </View>
  );
}
