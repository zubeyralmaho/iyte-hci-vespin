import { View } from "react-native";
import { AppText } from "@/components/ui/AppText";
import { Button } from "@/components/ui/Button";
import { Icon } from "@/components/ui/Icon";

type Props = {
  message?: string;
  onRetry?: () => void;
  fullScreen?: boolean;
};

export function ErrorMessage({
  message = "Something went wrong. Please try again.",
  onRetry,
  fullScreen = true,
}: Props) {
  const content = (
    <View className="items-center gap-3 px-6">
      <Icon name="alert-circle-outline" size="lg" tone="danger" />
      <AppText variant="body" tone="danger" className="text-center">
        {message}
      </AppText>
      {onRetry ? (
        <Button
          label="Retry"
          variant="secondary"
          onPress={onRetry}
          className="mt-2"
        />
      ) : null}
    </View>
  );

  if (fullScreen) {
    return (
      <View className="flex-1 items-center justify-center bg-background">
        {content}
      </View>
    );
  }

  return content;
}
