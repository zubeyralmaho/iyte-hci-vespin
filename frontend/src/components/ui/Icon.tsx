import { Ionicons } from "@expo/vector-icons";
import { semantic } from "@/theme/colors";
import { useSurfaceTone } from "@/providers/SurfaceProvider";

export type IconName = React.ComponentProps<typeof Ionicons>["name"];
export type IconSize = "sm" | "md" | "lg";
export type IconTone = "default" | "muted" | "onPrimary" | "danger";

const SIZE: Record<IconSize, number> = {
  sm: 16,
  md: 22,
  lg: 28,
};

const TONE: Record<IconTone, string> = {
  default: semantic.ink,
  muted: semantic.muted,
  onPrimary: semantic.onPrimary,
  danger: semantic.danger,
};

type Props = {
  name: IconName;
  size?: IconSize;
  tone?: IconTone;
};

export function Icon({ name, size = "md", tone }: Props) {
  const surfaceTone = useSurfaceTone();
  const resolvedTone: IconTone = tone ?? (surfaceTone === "primary" ? "onPrimary" : "default");
  return <Ionicons name={name} size={SIZE[size]} color={TONE[resolvedTone]} />;
}
