import { CardBehaviorDto } from "@/types/generated/api-types.ts";
import { ClassifiedBehavior } from "../types.ts";

export const classifyBehaviors = (behaviors: CardBehaviorDto[]): ClassifiedBehavior[] => {
  return behaviors.flatMap((behavior, originalIndex): ClassifiedBehavior | ClassifiedBehavior[] => {
    const hasTrigger = behavior.triggers && behavior.triggers.length > 0;
    const triggerType = hasTrigger ? behavior.triggers?.[0]?.type : null;
    const hasCondition = behavior.triggers?.[0]?.condition !== undefined;
    const hasInputs = behavior.inputs && behavior.inputs.length > 0;
    const hasProduction =
      behavior.outputs &&
      behavior.outputs.some((output: any) => output.type?.includes("production"));

    const hasDiscount =
      behavior.outputs && behavior.outputs.some((output: any) => output.type === "discount");

    const hasPaymentSubstitute =
      behavior.outputs &&
      behavior.outputs.some((output: any) => output.type === "payment-substitute");

    const hasValueModifier =
      behavior.outputs && behavior.outputs.some((output: any) => output.type === "value-modifier");

    const hasDefense =
      behavior.outputs && behavior.outputs.some((output: any) => output.type === "defense");

    const { description } = behavior;

    if (hasDiscount) {
      const discountOutputs = behavior.outputs?.filter((o: any) => o.type === "discount") ?? [];
      const otherOutputs = behavior.outputs?.filter((o: any) => o.type !== "discount") ?? [];

      if (otherOutputs.length > 0) {
        const discountBehavior = { ...behavior, outputs: discountOutputs };
        const remainderBehavior = { ...behavior, outputs: otherOutputs };
        return [
          { behavior: discountBehavior, type: "discount" as const, description, originalIndex },
          {
            behavior: remainderBehavior,
            type: "auto-no-background" as const,
            description: undefined,
            originalIndex,
          },
        ];
      }

      return [{ behavior, type: "discount" as const, description, originalIndex }];
    }

    if (hasPaymentSubstitute) {
      return { behavior, type: "payment-substitute", description, originalIndex };
    }

    if (hasValueModifier) {
      return { behavior, type: "value-modifier", description, originalIndex };
    }

    if (hasDefense) {
      return { behavior, type: "defense", description, originalIndex };
    }

    if (triggerType === "manual") {
      return { behavior, type: "manual-action", description, originalIndex };
    }

    if (triggerType === "auto" && hasCondition) {
      return { behavior, type: "triggered-effect", description, originalIndex };
    }

    if (triggerType === "auto" && !hasInputs) {
      return { behavior, type: "auto-no-background", description, originalIndex };
    }

    if (hasTrigger && hasInputs) {
      return { behavior, type: "triggered-effect", description, originalIndex };
    }

    if (hasProduction && (!hasTrigger || triggerType === "auto")) {
      return { behavior, type: "immediate-production", description, originalIndex };
    }

    return { behavior, type: "immediate-effect", description, originalIndex };
  });
};
