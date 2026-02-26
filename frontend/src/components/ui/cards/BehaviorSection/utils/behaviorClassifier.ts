import { CardBehaviorDto } from "@/types/generated/api-types.ts";
import { ClassifiedBehavior } from "../types.ts";

export const classifyBehaviors = (behaviors: CardBehaviorDto[]): ClassifiedBehavior[] => {
  return behaviors.flatMap((behavior): ClassifiedBehavior | ClassifiedBehavior[] => {
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
          { behavior: discountBehavior, type: "discount" as const, description },
          {
            behavior: remainderBehavior,
            type: "auto-no-background" as const,
            description: undefined,
          },
        ];
      }

      return [{ behavior, type: "discount" as const, description }];
    }

    if (hasPaymentSubstitute) {
      return { behavior, type: "payment-substitute", description };
    }

    if (hasValueModifier) {
      return { behavior, type: "value-modifier", description };
    }

    if (hasDefense) {
      return { behavior, type: "defense", description };
    }

    if (triggerType === "manual") {
      return { behavior, type: "manual-action", description };
    }

    if (triggerType === "auto" && hasCondition) {
      return { behavior, type: "triggered-effect", description };
    }

    if (triggerType === "auto" && !hasInputs) {
      return { behavior, type: "auto-no-background", description };
    }

    if (hasTrigger && hasInputs) {
      return { behavior, type: "triggered-effect", description };
    }

    if (hasProduction && (!hasTrigger || triggerType === "auto")) {
      return { behavior, type: "immediate-production", description };
    }

    return { behavior, type: "immediate-effect", description };
  });
};
