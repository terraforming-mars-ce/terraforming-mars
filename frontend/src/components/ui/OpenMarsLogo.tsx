import { assetImage } from "@/assets";

interface LogoProps {
  surface: "menu" | "loading";
}

export default function OpenMarsLogo({ surface }: LogoProps) {
  return (
    <span
      role="img"
      aria-label="Open Mars"
      className="inline-flex w-full max-w-[520px] flex-col items-center justify-center text-[#eaf1f7] compact:max-w-[340px]"
    >
      <FrontierLogo surface={surface} />
    </span>
  );
}

function FrontierLogo({ surface }: LogoProps) {
  if (surface === "loading") {
    return (
      <span aria-hidden="true" className="block aspect-[3/1] w-full overflow-hidden">
        <img
          {...assetImage("branding/red-frontier", "(max-width: 600px) 340px, 520px")}
          alt=""
          width={1536}
          height={1024}
          className="block h-full w-full object-cover object-[50%_43%]"
          fetchPriority="high"
        />
      </span>
    );
  }

  return (
    <span aria-hidden="true" className="block w-full [container-type:inline-size]">
      <span className="flex flex-col items-center gap-[3cqw]">
        <span className="font-orbitron flex w-[56%] justify-between text-[5.5cqw] leading-none font-normal">
          {Array.from("OPEN").map((letter) => (
            <span key={letter}>{letter}</span>
          ))}
        </span>
        <span className="font-frontier block bg-linear-to-b from-[#f2f5f8] to-[#aab9c7] bg-clip-text text-[25cqw] leading-[0.8] text-transparent">
          MARS
        </span>
      </span>
    </span>
  );
}
