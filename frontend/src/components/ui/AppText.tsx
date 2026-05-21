import { Text, type TextProps } from "react-native";
import { useSurfaceTone, type SurfaceTone } from "@/providers/SurfaceProvider";

export type TextVariant = "display" | "title" | "body" | "button" | "caption";
export type TextTone = "default" | "muted" | "onPrimary" | "danger";

const VARIANT: Record<TextVariant, string> = {
  display: "font-display text-[40px] leading-[44px]",
  button: "font-display text-base leading-5",
  title: "text-xl leading-7 font-semibold",
  body: "text-base leading-6",
  caption: "text-[13px] leading-[18px]",
};

const TONE: Record<TextTone, string> = {
  default: "text-ink",
  muted: "text-muted",
  onPrimary: "text-onPrimary",
  danger: "text-danger",
};

const SURFACE_DEFAULT_TONE: Record<SurfaceTone, TextTone> = {
  default: "default",
  primary: "onPrimary",
};

export function resolveTextClasses(
  variant: TextVariant,
  tone: TextTone | undefined,
  surfaceTone: SurfaceTone,
  className?: string,
): string {
  const resolvedTone = tone ?? SURFACE_DEFAULT_TONE[surfaceTone];
  return [VARIANT[variant], TONE[resolvedTone], className].filter(Boolean).join(" ");
}

type Props = TextProps & {
  variant?: TextVariant;
  tone?: TextTone;
  className?: string;
};

export function AppText({
  variant = "body",
  tone,
  className,
  ...rest
}: Props) {
  const surfaceTone = useSurfaceTone();
  return <Text className={resolveTextClasses(variant, tone, surfaceTone, className)} {...rest} />;
}
