// Corporation logos from TM Cards List
import React from "react";
import "./corporationLogos.css";

export const corporationLogos: Record<string, React.ReactNode> = {
  credicor: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "40px",
          color: "purple",
          fontFamily: "'Times New Roman'",
          fontWeight: "normal",
          lineHeight: "60px",
          border: "2px solid purple",
          paddingLeft: "5px",
          paddingBottom: "5px",
          paddingRight: "5px",
        }}
      >
        credicor
      </span>
    </div>
  ),
  ecoline: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "50px",
          fontWeight: "normal",
          color: "rgb(0, 180, 0)",
          letterSpacing: "2px",
        }}
      >
        ecoline
      </span>
    </div>
  ),
  helion: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "34px",
          width: "140px",
          textAlign: "center",
          border: "2px solid black",
          background: "#e6e600",
          borderRadius: "2px",
          color: "black",
        }}
      >
        helion
      </div>
    </div>
  ),
  "mining-guild": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        className="mining guild"
        style={{
          fontSize: "24px",
          color: "#c9380e",
          display: "inline-block",
          WebkitTransform: "scale(1.5, 1)",
          MozTransform: "scale(1.5, 1)",
          msTransform: "scale(1.5, 1)",
          OTransform: "scale(1.5, 1)",
          transform: "scale(1.5, 1)",
        }}
      >
        MINING
        <br />
        GUILD
      </span>
    </div>
  ),
  "interplanetary-cinematics": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="INTERPLANETARY CINEMATICS"
        style={{
          fontSize: "17px",
          color: "white",
        }}
      >
        INTERPLANETARY
      </div>
      <div
        style={{
          height: "5px",
          width: "143px",
          border: "5px solid #cc3333",
        }}
      ></div>
      <div
        className="INTERPLANETARY CINEMATICS"
        style={{
          fontSize: "24px",
          display: "inline-block",
          WebkitTransform: "scale(0.5, 1)",
          MozTransform: "scale(0.5, 1)",
          msTransform: "scale(0.5, 1)",
          OTransform: "scale(0.5, 1)",
          transform: "scale(1, 0.5)",
          color: "grey",
        }}
      >
        CINEMATICS
      </div>
    </div>
  ),
  inventrix: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "24px",
          paddingLeft: "5px",
          paddingBottom: "5px",
          color: "grey",
          whiteSpace: "nowrap",
        }}
      >
        <span
          style={{
            backgroundColor: "#6bb5c7",
            paddingLeft: "4px",
            paddingRight: "4px",
            fontSize: "26px",
            color: "black",
            marginRight: "5px",
          }}
        >
          X
        </span>
        INVENTRIX
      </span>
    </div>
  ),
  phobolog: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "24px",
          color: "white",
          lineHeight: "40px",
          background: "#32004d",
          paddingLeft: "5px",
          paddingRight: "5px",
          border: "1px solid #444",
          borderRadius: "10px",
          fontFamily: "'Times New Roman'",
          display: "inline-block",
          WebkitTransform: "scale(1.2, 1)",
          MozTransform: "scale(1.2, 1)",
          msTransform: "scale(1.2, 1)",
          OTransform: "scale(1.2, 1)",
          transform: "scale(1.2, 1)",
        }}
      >
        PHOBOLOG
      </span>
    </div>
  ),
  "tharsis-republic": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "24px",
          color: "lightgrey",
          whiteSpace: "nowrap",
        }}
      >
        <div
          style={{
            textShadow: "none",
            display: "inline-block",
            backgroundColor: "#ff5f00",
            marginRight: "5px",
            color: "white",
          }}
        >
          <span>▲</span>
          <span
            style={{
              fontSize: "14px",
              padding: "0px",
              border: "none",
            }}
          >
            ▲
          </span>
        </div>
        THARSIS
        <br />
        &nbsp; REPUBLIC
      </div>
    </div>
  ),
  thorgate: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "32px",
          fontFamily: "'Arial Narrow', 'Verdana'",
          fontWeight: "normal",
          color: "grey",
        }}
      >
        THORGATE
      </span>
    </div>
  ),
  "united-nations-mars-initiative": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="background-color-active"
        style={{
          fontSize: "16px",
          width: "100px",
          color: "white",
          padding: "5px",
          paddingTop: "10px",
          textAlign: "center",
          fontWeight: "normal",
          textShadow: "0 0 2px black",
        }}
      >
        UNITED NATIONS MARS INITIATIVE
      </div>
    </div>
  ),
  teractor: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "34px",
          color: "orangered",
          fontFamily: "'Times New Roman'",
          fontWeight: "normal",
        }}
      >
        TERACTOR
      </span>
    </div>
  ),
  "saturn-systems": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "14px",
          color: "white",
          lineHeight: "40px",
          background: "#32004d",
          paddingTop: "8px",
          paddingBottom: "8px",
          paddingLeft: "20px",
          paddingRight: "20px",
          borderRadius: "50%",
          fontWeight: "normal",
          border: "2px solid white",
          whiteSpace: "nowrap",
        }}
      >
        SATURN
        <span
          style={{
            fontSize: "20px",
            display: "inline-block",
          }}
        >
          ●
        </span>
        SYSTEMS
      </span>
    </div>
  ),
  aphrodite: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "23px",
          color: "orange",
          fontWeight: "bold",
        }}
      >
        APHRODITE
      </div>
    </div>
  ),
  celestic: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="celestic"
        style={{
          fontSize: "24px",
          width: "100px",
          color: "black",
        }}
      >
        <span
          style={{
            background:
              "linear-gradient( to right, rgb(251, 192, 137), rgb(251, 192, 137), rgb(23, 185, 236) )",
            paddingLeft: "5px",
          }}
        >
          CEL
        </span>
        <span
          style={{
            background: "linear-gradient( to right, rgb(23, 185, 236), rgb(251, 192, 137) )",
          }}
        >
          ES
        </span>
        <span
          style={{
            background: "rgb(251, 192, 137)",
            paddingRight: "5px",
          }}
        >
          TIC
        </span>
      </div>
    </div>
  ),
  manutech: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        className="manutech"
        style={{
          fontSize: "30px",
          color: "#e63900",
        }}
      >
        <span
          style={{
            color: "white",
            background: "#e63900",
            textShadow: "none",
            paddingLeft: "2px",
          }}
        >
          MA
        </span>
        NUTECH
      </span>
    </div>
  ),
  "morning-star-inc": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "18px",
          color: "white",
        }}
      >
        MORNING STAR INC.
      </div>
    </div>
  ),
  viron: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "50px",
          fontFamily: "Prototype",
          color: "grey",
        }}
      >
        VIRON
      </div>
    </div>
  ),
  "cheung-shing-mars": (
    <div
      style={{
        display: "flex",
        flexDirection: "row",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          width: "50px",
          display: "inline-block",
          marginTop: "10px",
          marginRight: "-10px",
        }}
      >
        <span
          style={{
            color: "red",
            border: "4px solid red",
            borderRadius: "50%",
            padding: "3px 5px 3px 5px",
            fontSize: "30px",
            lineHeight: "14px",
          }}
        >
          㨐
        </span>
      </div>
      <div
        style={{
          display: "inline-block",
          width: "140px",
          fontSize: "19px",
          lineHeight: "22px",
          verticalAlign: "middle",
          fontFamily: "'Prototype'",
          fontWeight: "normal",
          color: "lightgray",
        }}
      >
        &nbsp;Cheung Shing
        <br />
        <div style={{}}>■■MARS■■</div>
      </div>
    </div>
  ),
  "point-luna": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="point luna"
        style={{
          fontSize: "22px",
          fontFamily: "Prototype",
          fontWeight: "normal",
          display: "inline-block",
          textDecoration: "underline",
          WebkitTransform: "scale(1.5, 1)",
          MozTransform: "scale(1.5, 1)",
          msTransform: "scale(1.5, 1)",
          OTransform: "scale(1.5, 1)",
          transform: "scale(1.5, 1)",
          color: "white",
        }}
      >
        POINT<span>&nbsp;</span>LUNA
      </div>
    </div>
  ),
  "robinson-industries": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="robinson"
        style={{
          letterSpacing: "4px",
          borderBottom: "3px solid #ccc",
          color: "black",
          padding: "4px",
        }}
      >
        ROBINSON
      </div>
      <div
        className="robinson"
        style={{
          borderBottom: "3px solid #ccc",
          color: "black",
          padding: "4px",
        }}
      >
        •—•—•—•—•—•—•&nbsp;
      </div>
      <div
        className="robinson"
        style={{
          letterSpacing: "2px",
          color: "black",
          padding: "4px",
        }}
      >
        INDUSTRIES
      </div>
    </div>
  ),
  "valley-trust": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          color: "rgb(2, 125, 195)",
          background:
            "linear-gradient( to right, rgb(2, 125, 195) 10%, white, white, white, white, white, white, white )",
          boxShadow: "3px 3px 10px 1px rgb(58, 58, 58)",
          width: "135px",
          lineHeight: "24px",
          borderRadius: "10px 0px 0px 10px",
        }}
      >
        <div
          style={{
            display: "inline-block",
            fontSize: "26px",
            textAlign: "center",
            color: "black",
            fontWeight: "bold",
          }}
        >
          VALLEY TRUST
        </div>
      </div>
    </div>
  ),
  vitor: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="vitor"
        style={{
          fontSize: "24px",
          display: "inline-block",
          WebkitTransform: "scale(2, 1)",
          MozTransform: "scale(2, 1)",
          msTransform: "scale(2, 1)",
          OTransform: "scale(2, 1)",
          transform: "scale(2, 1)",
          color: "black",
        }}
      >
        <span
          style={{
            color: "white",
            background: "orangered",
            paddingLeft: "3px",
          }}
        >
          VIT
        </span>
        <span
          style={{
            background: "linear-gradient(to right, orangered, white)",
          }}
        >
          O
        </span>
        <span
          style={{
            background: "white",
            paddingRight: "3px",
          }}
        >
          R
        </span>
      </div>
    </div>
  ),
  aridor: <div className="aridor">ARIDOR</div>,
  arklight: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="arklight"
        style={{
          fontSize: "19px",
          fontFamily: "Prototype",
          letterSpacing: "1px",
          padding: "4px",
          background: "linear-gradient( to right, #000089, lightblue )",
          WebkitTransform: "scale(2, 1)",
          MozTransform: "scale(2, 1)",
          msTransform: "scale(2, 1)",
          OTransform: "scale(2, 1)",
          transform: "scale(2, 1)",
        }}
      >
        ARKLIGHT
      </div>
    </div>
  ),
  polyphemos: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="polyphemos">
        <span className="polyphemos2">POL</span>YPHEMOS
      </div>
    </div>
  ),
  poseidon: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="poseidon">POSEIDON</div>
    </div>
  ),
  "stormcraft-incorporated": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div style={{ display: "flex" }}>
        <div className="stormcraft1">STORM</div>
        <div className="stormcraft2">CRAFT</div>
      </div>
      <div style={{ display: "flex", gap: "34px" }}>
        <div className="stormcraft3">INCOR</div>
        <div className="stormcraft4">PORATED</div>
      </div>
    </div>
  ),
  "lakefront-resorts": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "22px",
          fontFamily: "Times",
          color: "white",
          textShadow: "0 1px 0px #444, 0px -1px 0px #444, -1px 0px 0px #444, 1px 0px 0px #444",
          letterSpacing: "4px",
        }}
      >
        LAKEFRONT
        <br />
        &nbsp;&nbsp;RESORTS
      </div>
    </div>
  ),
  pristar: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          color: "#ff5d21",
          textShadow: "3px 3px 3px black",
          fontSize: "30px",
          transform: "scaleX(0.8)",
          letterSpacing: "10px",
        }}
      >
        PRISTAR
      </div>
    </div>
  ),
  "septem-tribus": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="septem">Septem Tribus</div>
    </div>
  ),
  "terralabs-research": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "16px",
          fontFamily: "Prototype",
          color: "white",
          transform: "scale(2, 1)",
        }}
      >
        TERRALABS
      </div>
      <div
        style={{
          fontSize: "8px",
          letterSpacing: "2px",
          fontFamily: "Prototype",
          transform: "scale(2, 1)",
          color: "lightgrey",
        }}
      >
        RESEARCH
      </div>
    </div>
  ),
  "utopia-invest": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="utopia">
        <div className="utopia1">UTOPIA</div>
        <div className="utopia2">INVEST</div>
      </div>
    </div>
  ),
  factorum: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="factorum">FACTORUM</div>
    </div>
  ),
  "mons-insurance": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="mons">
        <div className="mons0">▲</div>
        <div className="mons1">mons</div>
        <div className="mons2">INSURANCE</div>
      </div>
    </div>
  ),
  philares: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="philares">
        PHIL
        <span
          style={{
            color: "#ff5858",
          }}
        >
          A
        </span>
        RES
      </div>
    </div>
  ),
  "arcadian-communities": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "20px",
          paddingLeft: "3px",
          width: "157px",
          background: "#eeeeee",
          boxShadow: "0 0 0 1px rgba(0, 0, 0, 0.6), 3px 3px 3px grey",
          borderRadius: "5px",
          borderTop: "2px solid rgb(221, 221, 221)",
          borderLeft: "2px solid rgb(221, 221, 221)",
          borderBottom: "2px solid rgb(137, 137, 137)",
          borderRight: "2px solid rgb(137, 137, 137)",
          fontFamily: "Arial",
          color: "black",
        }}
      >
        &nbsp;&nbsp;&nbsp;ARCADIAN
        <br />
        COMMUNITIES
      </div>
    </div>
  ),
  recyclon: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "20px",
          borderRadius: "25px",
          padding: "10px",
          background: "red",
          color: "white",
          boxShadow: "0 0 0 1px rgba(0, 0, 0, 0.6), 0 0 0 2px rgba(0, 0, 0, 0.3), 3px 3px 3px #444",
          fontFamily: "Prototype",
          fontWeight: "normal",
          width: "120px",
        }}
      >
        RECYCLON
      </div>
    </div>
  ),
  "splice-tactical-genomics": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className=""
        style={{
          fontSize: "29px",
          fontFamily: "Arial",
          fontWeight: "bold",
          width: "109px",
          background: "#eeeeee",
          color: "black",
          paddingLeft: 2,
          paddingRight: 2,
        }}
      >
        <div style={{}}>
          SPLI
          <span
            style={{
              color: "red",
            }}
          >
            C
          </span>
          E
        </div>
        <div
          style={{
            height: "3px",
            background: "red",
          }}
        ></div>
        <div
          style={{
            fontSize: "10px",
          }}
        >
          TACTICAL GENOMICS
        </div>
      </div>
    </div>
  ),
  astrodrill: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div className="astrodrill">ASTRODRILL</div>
      <div className="astrodrill" style={{}}></div>
    </div>
  ),
  "pharmacy-union": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          fontSize: "16px",
        }}
      >
        PHARMACY UNION
      </div>
    </div>
  ),
  ecotec: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        className="ecotec"
        style={{
          fontSize: "30px",
          paddingBottom: "5px",
          fontWeight: "normal",
          color: "yellow",
          letterSpacing: "2px",
          backgroundColor: "green",
          paddingLeft: "5px",
          borderRadius: "5px 0 0 5px",
        }}
      >
        eco
      </span>
      <span
        style={{
          fontSize: "30px",
          paddingBottom: "5px",
          fontWeight: "normal",
          color: "green",
          letterSpacing: "2px",
          backgroundColor: "yellow",
          paddingRight: "5px",
          borderRadius: "0px 5px 5px 0px",
        }}
      >
        tec
      </span>
    </div>
  ),
  "tycho-magnetics": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <span
        style={{
          fontSize: "32px",
          fontFamily: "'Arial Narrow', 'Verdana'",
          fontWeight: "normal",
          border: "1px solid black",
          borderRight: "none",
          paddingLeft: "5px",
          color: "grey",
        }}
      >
        TYCHO
      </span>
      <br />
      <span
        style={{
          border: "1px solid black",
          borderRight: "none",
          paddingLeft: "5px",
          color: "grey",
        }}
      >
        MAGNETICS
      </span>
    </div>
  ),
  "kuiper-cooperative": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          backgroundColor: "rgb(98, 126, 219)",
          width: "120px",
          border: "2px solid #aaa",
          borderBottom: "5px solid rgb(8, 8, 88)",
          paddingLeft: "5px",
          color: "white",
          letterSpacing: "1px",
          fontSize: "16px",
        }}
      >
        KUIPER
      </div>
      <div
        style={{
          width: "120px",
          paddingLeft: "5px",
          color: "#cc3333",
          letterSpacing: "1px",
          fontSize: "16px",
          border: "2px solid rgb(146, 163, 219)",
          borderTop: "none",
        }}
      >
        COOPERATIVE
      </div>
    </div>
  ),
  spire: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          color: "white",
          letterSpacing: "1px",
          fontSize: "16px",
          paddingTop: "15px",
          transform: "scaleX(1.5)",
          width: "80px",
          textAlign: "center",
          border: "1px solid white",
          boxShadow: "0 0 0 2px black",
          background: "linear-gradient(#222 40%, #aaa 40%, black 90%)",
          fontFamily: "'Times New Roman', Times, serif",
        }}
      >
        SPIRE
      </div>
    </div>
  ),
  sagitta: (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          color: "rgb(82, 190, 82)",
          textShadow: "0 0 2px black",
          fontFamily: "Georgia, 'Times New Roman', Times, serif, sans-serif",
          letterSpacing: "5px",
          fontSize: "30px",
        }}
      >
        SAGITTA
      </div>
    </div>
  ),
  "palladin-shipping": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        style={{
          boxShadow: "0 2px 5px 2px black",
          background: "rgb(218, 218, 115)",
          borderRadius: "40%",
          width: "150px",
          color: "#444",
          letterSpacing: "2px",
          textAlign: "center",
        }}
      >
        <div style={{}}>PALLADIN</div>
        <div style={{}}>SHIPPING</div>
      </div>
    </div>
  ),
  "nirgal-enterprises": (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "4px",
      }}
    >
      <div
        className="NIRGAL"
        style={{
          color: "rgb(226, 153, 17)",
          background: "linear-gradient(transparent, white)",
          letterSpacing: "2px",
          borderBottom: "2px solid rgb(35, 35, 160)",
          width: "190px",
          textAlign: "center",
          paddingBottom: "3px",
        }}
      >
        Nirgal Enterprises
      </div>
    </div>
  ),
};

export function getCorporationLogo(
  corporationName: keyof typeof corporationLogos,
): React.ReactNode | null {
  const key = corporationName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  return corporationLogos[key] || null;
}
