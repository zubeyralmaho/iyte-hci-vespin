import { Image } from "react-native";

const MARK_SOURCE = require("../../../assets/brand/mark.png");

export type MarkSize = "sm" | "md" | "lg";

const SIZE: Record<MarkSize, number> = {
  sm: 48,
  md: 96,
  lg: 160,
};

type Props = {
  size?: MarkSize;
};

export function Mark({ size = "md" }: Props) {
  const px = SIZE[size];
  return (
    <Image
      source={MARK_SOURCE}
      style={{ width: px, height: px }}
      resizeMode="contain"
      accessibilityRole="image"
      accessibilityLabel="Vespin"
    />
  );
}
