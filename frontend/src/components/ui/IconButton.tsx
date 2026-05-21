import { Pressable, type PressableProps } from "react-native";
import { Icon, type IconName, type IconSize, type IconTone } from "./Icon";

type Props = Omit<PressableProps, "children" | "accessibilityLabel"> & {
  name: IconName;
  accessibilityLabel: string;
  size?: IconSize;
  tone?: IconTone;
  className?: string;
};

export function IconButton({
  name,
  accessibilityLabel,
  size = "md",
  tone,
  disabled,
  className,
  ...rest
}: Props) {
  const cls = [
    "h-11 w-11 items-center justify-center rounded-pill",
    disabled ? "opacity-50" : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={accessibilityLabel}
      disabled={disabled}
      className={cls}
      {...rest}
    >
      <Icon name={name} size={size} tone={tone} />
    </Pressable>
  );
}
