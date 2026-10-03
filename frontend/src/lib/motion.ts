import { motion } from "@chat/design-tokens/motion";

import type { MotionToken, MotionTokenName } from "@chat/design-tokens/motion";
import type { Transition } from "motion/react";

const toTransition = (token: MotionToken): Transition =>
  token.type === "timing"
    ? { duration: token.duration / 1000, ease: [...token.easing] }
    : { damping: token.damping, mass: token.mass, stiffness: token.stiffness, type: "spring" };

// Motion に渡す transition。値を部品に直接書かず、ここから選ぶ
export const transitions: Record<MotionTokenName, Transition> = {
  base: toTransition(motion.base),
  fast: toTransition(motion.fast),
  push: toTransition(motion.push),
  sheet: toTransition(motion.sheet),
  spring: toTransition(motion.spring),
};
