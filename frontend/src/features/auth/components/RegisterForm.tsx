import { zodResolver } from "@hookform/resolvers/zod";
import { Link } from "expo-router";
import { Controller, useForm } from "react-hook-form";
import { Pressable, Text, TextInput, View } from "react-native";
import { ApiError } from "@/api/error";
import { useTranslation } from "@/providers/I18nProvider";
import { useRegister } from "../hooks/useAuthActions";
import { registerSchema, type RegisterInput } from "../schemas/register";

export function RegisterForm() {
  const { t } = useTranslation();
  const register = useRegister();
  const form = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      displayName: "",
    },
  });

  const submit = form.handleSubmit((values) => register.mutate(values));
  const errorMessage =
    register.error instanceof ApiError ? register.error.message : register.error ? t("auth.errors.generic") : null;

  return (
    <View className="flex-1 justify-center bg-background p-6">
      <View className="gap-6">
        <View className="gap-2">
          <Text className="text-3xl font-semibold text-ink">{t("auth.register.title")}</Text>
          <Text className="text-base text-muted">{t("auth.register.subtitle")}</Text>
        </View>

        <View className="gap-4">
          <View className="gap-2">
            <Text className="text-sm font-medium text-ink">{t("auth.fields.displayName")}</Text>
            <Controller
              control={form.control}
              name="displayName"
              render={({ field }) => (
                <TextInput
                  autoCapitalize="words"
                  autoComplete="name"
                  onBlur={field.onBlur}
                  onChangeText={field.onChange}
                  value={field.value}
                  className="rounded-md border border-border bg-surface px-4 py-3 text-ink"
                />
              )}
            />
          </View>

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
                    autoComplete="new-password"
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
            disabled={register.isPending}
            onPress={submit}
            className="items-center rounded-md bg-primary px-4 py-3 disabled:opacity-60"
          >
            <Text className="font-semibold text-onPrimary">
              {register.isPending ? t("auth.register.submitting") : t("auth.register.submit")}
            </Text>
          </Pressable>
        </View>

        <Link href="/(auth)/login" asChild>
          <Pressable accessibilityRole="button">
            <Text className="text-center font-medium text-primary">{t("auth.register.loginLink")}</Text>
          </Pressable>
        </Link>
      </View>
    </View>
  );
}
