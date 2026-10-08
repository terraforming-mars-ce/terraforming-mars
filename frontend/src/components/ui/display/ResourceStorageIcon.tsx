import React from "react";
import { getIconPath } from "@/utils/iconStore.ts";
import { ResourceStorageDto } from "@/types/generated/api-types.ts";
import DecorBox from "./DecorBox.tsx";
import RevealTrigger from "./RevealTrigger.tsx";

interface ResourceStorageIconProps {
  resourceStorage?: ResourceStorageDto;
  corner?: "bottom-right" | "top-left";
  bare?: boolean;
}

const ResourceStorageIcon: React.FC<ResourceStorageIconProps> = ({
  resourceStorage,
  corner = "bottom-right",
  bare = false,
}) => {
  if (!resourceStorage) {
    return null;
  }

  const resourceIcon = getIconPath(resourceStorage.type);
  const label = (
    <>
      {resourceIcon && (
        <img src={resourceIcon} alt={resourceStorage.type} className="w-3.5 h-3.5 object-contain" />
      )}
      <span className="text-[9px] text-white/50 font-semibold tracking-wider uppercase">
        {resourceStorage.type}
      </span>
    </>
  );

  return (
    <RevealTrigger
      as="div"
      className="relative w-fit"
      content={resourceStorage.description || null}
      placement="below"
    >
      {bare ? (
        <div className="inline-flex items-center gap-1">{label}</div>
      ) : (
        <DecorBox corner={corner}>{label}</DecorBox>
      )}
    </RevealTrigger>
  );
};

export default ResourceStorageIcon;
