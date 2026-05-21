import { Pressable, type PressableProps } from "react-native";
import { AppText, type TextTone } from "./AppText";

export type ButtonVariant = "primary" | "secondary" | "ghost";
export type ButtonSize = "md" | "lg";

const BTN_VARIANT: Record<ButtonVariant, string> = {
  primary: "bg-primary",
  secondary: "bg-surface border border-border",
  ghost: "bg-transparent",
};

const BTN_SIZE: Record<ButtonSize, string> = {
  md: "h-11 px-4 rounded-pill",
  lg: "h-14 px-6 rounded-pill",
};

const DEFAULT_LABEL_TONE: Record<ButtonVariant, TextTone> = {
  primary: "onPrimary",
  secondary: "default",
  ghost: "default",
};

type Props = Omit<PressableProps, "children"> & {
  label: string;
  variant?: ButtonVariant;
  size?: ButtonSize;
  labelTone?: TextTone;
  fullWidth?: boolean;
  className?: string;
};

export function Button({
  label,
  variant = "primary",
  size = "md",
  labelTone,
  fullWidth,
  disabled,
  className,
  ...rest
}: Props) {
  const tone = labelTone ?? DEFAULT_LABEL_TONE[variant];
  const cls = [
    "items-center justify-center",
    BTN_VARIANT[variant],
    BTN_SIZE[size],
    fullWidth ? "w-full" : "",
    disabled ? "opacity-50" : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <Pressable
      accessibilityRole="button"
      disabled={disabled}
      className={cls}
      {...rest}
    >
      <AppText variant="button" tone={tone}>
        {label}
      </AppText>
    </Pressable>
  );
}
