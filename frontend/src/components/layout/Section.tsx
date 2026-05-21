import { View, type ViewProps } from "react-native";
import { AppText } from "@/components/ui/AppText";

type Props = ViewProps & {
  title?: string;
  className?: string;
};

export function Section({ title, className, children, ...rest }: Props) {
  return (
    <View className={["gap-3", className].filter(Boolean).join(" ")} {...rest}>
      {title ? <AppText variant="title">{title}</AppText> : null}
      {children}
    </View>
  );
}
