import { useEffect, useRef } from "react";
import { Animated, View } from "react-native";
import { AppText } from "@/components/ui/AppText";
import { Icon, type IconName } from "@/components/ui/Icon";
import type { IconTone } from "@/components/ui/Icon";

export type ToastVariant = "success" | "error" | "info";

const VARIANT_CONFIG: Record<ToastVariant, { icon: IconName; tone: IconTone; bg: string }> = {
  success: { icon: "checkmark-circle-outline", tone: "default", bg: "bg-surface border border-border" },
  error: { icon: "alert-circle-outline", tone: "danger", bg: "bg-surface border border-danger" },
  info: { icon: "information-circle-outline", tone: "muted", bg: "bg-surface border border-border" },
};

type Props = {
  message: string;
  variant?: ToastVariant;
  visible: boolean;
  duration?: number;
  onDismiss?: () => void;
};

export function Toast({
  message,
  variant = "info",
  visible,
  duration = 3000,
  onDismiss,
}: Props) {
  const opacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    if (visible) {
      Animated.timing(opacity, {
        toValue: 1,
        duration: 200,
        useNativeDriver: true,
      }).start();

      if (onDismiss && duration > 0) {
        const timer = setTimeout(onDismiss, duration);
        return () => clearTimeout(timer);
      }
    } else {
      Animated.timing(opacity, {
        toValue: 0,
        duration: 200,
        useNativeDriver: true,
      }).start();
    }
  }, [visible, duration, onDismiss, opacity]);

  if (!visible) return null;

  const config = VARIANT_CONFIG[variant];

  return (
    <Animated.View style={{ opacity }} className="absolute bottom-6 left-4 right-4 z-50">
      <View className={`flex-row items-center gap-3 rounded-pill px-4 py-3 ${config.bg}`}>
        <Icon name={config.icon} size="md" tone={config.tone} />
        <AppText variant="body" className="flex-1 flex-shrink">
          {message}
        </AppText>
      </View>
    </Animated.View>
  );
}
