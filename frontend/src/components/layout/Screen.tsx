import { View, type ViewProps } from "react-native";
import { SafeAreaView, type Edge } from "react-native-safe-area-context";
import { SurfaceProvider, type SurfaceTone } from "@/providers/SurfaceProvider";

const TONE_BG: Record<SurfaceTone, string> = {
  default: "bg-background",
  primary: "bg-primary",
};

type Props = ViewProps & {
  tone?: SurfaceTone;
  edges?: readonly Edge[];
  padded?: boolean;
  className?: string;
};

export function Screen({
  tone = "default",
  edges = ["top", "bottom"],
  padded = true,
  className,
  children,
  ...rest
}: Props) {
  const cls = [
    "flex-1",
    TONE_BG[tone],
    padded ? "px-5" : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <SurfaceProvider tone={tone}>
      <SafeAreaView edges={edges} className={`flex-1 ${TONE_BG[tone]}`}>
        <View className={cls} {...rest}>
          {children}
        </View>
      </SafeAreaView>
    </SurfaceProvider>
  );
}
