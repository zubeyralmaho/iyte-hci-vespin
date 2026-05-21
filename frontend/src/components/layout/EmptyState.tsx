import { View } from "react-native";
import { AppText } from "@/components/ui/AppText";

type Props = {
  title: string;
  description?: string;
  action?: React.ReactNode;
};

export function EmptyState({ title, description, action }: Props) {
  return (
    <View className="flex-1 items-center justify-center gap-3 px-6">
      <AppText variant="title" className="text-center">
        {title}
      </AppText>
      {description ? (
        <AppText variant="body" tone="muted" className="text-center">
          {description}
        </AppText>
      ) : null}
      {action}
    </View>
  );
}
