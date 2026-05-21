import { ActivityIndicator, View } from "react-native";
import { semantic } from "@/theme/colors";

type Props = {
  size?: "small" | "large";
  fullScreen?: boolean;
};

export function Spinner({ size = "large", fullScreen = true }: Props) {
  if (fullScreen) {
    return (
      <View className="flex-1 items-center justify-center bg-background">
        <ActivityIndicator size={size} color={semantic.primary} />
      </View>
    );
  }

  return <ActivityIndicator size={size} color={semantic.primary} />;
}
