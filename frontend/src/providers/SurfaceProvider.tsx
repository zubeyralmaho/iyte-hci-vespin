import { createContext, useContext, type ReactNode } from "react";

export type SurfaceTone = "default" | "primary";

const SurfaceContext = createContext<SurfaceTone>("default");

type Props = {
  tone: SurfaceTone;
  children: ReactNode;
};

export function SurfaceProvider({ tone, children }: Props) {
  return <SurfaceContext.Provider value={tone}>{children}</SurfaceContext.Provider>;
}

export function useSurfaceTone(): SurfaceTone {
  return useContext(SurfaceContext);
}
