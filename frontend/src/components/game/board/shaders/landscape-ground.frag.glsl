uniform sampler2D uMars;
uniform sampler2D uDryGrass;
uniform sampler2D uLushGrass;
uniform sampler2D uForestFloor;
uniform sampler2D uWetSoil;
uniform sampler2D uLushDetail;
uniform sampler2D uLitterDetail;
uniform sampler2D uSoilDetail;
uniform sampler2D uSand;
uniform sampler2D uRock;
uniform sampler2D uPaving;
uniform vec2 uBeachWidths;
varying vec2 vBoardPos;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
//#pragma body
vec4 field=landscapeField(vFieldLocal,vFieldLayer,vFieldBirth);
vec4 weights=landscapeMaterials(vFieldLocal,vFieldLayer,vFieldBirth);
float shore=field.y;
vec3 sphereDirection=normalize(landscapeProject(vBoardPos,0.0));
vec2 marsUv=vec2(atan(sphereDirection.z,-sphereDirection.x)/(2.0*PI),acos(clamp(sphereDirection.y,-1.0,1.0))/PI);
marsUv.y=1.0-marsUv.y;
vec3 base=texture2D(uMars,marsUv).rgb;
vec2 uv=vBoardPos*80.0;
vec2 materialUv=vBoardPos*18.0;
vec2 lushUv=mat2(0.8,-0.6,0.6,0.8)*materialUv*0.83;
vec2 litterUv=materialUv*0.7;
vec2 soilUv=materialUv*0.65;
vec4 lushDetail=texture2D(uLushDetail,lushUv);
vec4 litterDetail=texture2D(uLitterDetail,litterUv);
vec4 soilDetail=texture2D(uSoilDetail,soilUv);
vec3 dry=texture2D(uDryGrass,uv).rgb*vec3(0.48,0.66,0.32);
vec3 lush=texture2D(uLushGrass,lushUv).rgb*vec3(0.65,1.1,0.6);
vec3 litter=texture2D(uForestFloor,litterUv).rgb*vec3(0.48,0.58,0.4);
vec3 mud=texture2D(uWetSoil,soilUv).rgb*vec3(0.65,0.7,0.6);
vec4 materialWeights=weights*vec4(1.0,0.85+0.3*lushDetail.a,0.85+0.3*litterDetail.a,0.85+0.3*soilDetail.a);
materialWeights/=max(0.0001,dot(materialWeights,vec4(1.0)));
vec3 nature=dry*materialWeights.x+lush*materialWeights.y+litter*materialWeights.z+mud*materialWeights.w;
float coverage=clamp(field.z,0.0,1.0);
float growth=mix(coverage,1.0-pow(1.0-coverage,2.2),smoothstep(0.2,0.7,coverage));
float foundation=field.w;
float bankWidth=mix(uBeachWidths.y,uBeachWidths.x,clamp(weights.y+weights.z,0.0,1.0));
float bank=1.0-smoothstep(bankWidth*0.55,bankWidth,max(shore,0.0));
vec3 sand=texture2D(uSand,uv*0.6).rgb;
vec3 rock=texture2D(uRock,uv*0.37).rgb;
vec3 paving=texture2D(uPaving,vBoardPos*9.0).rgb;
float aggregate=dot(paving,vec3(0.2126,0.7152,0.0722));
vec3 preparedGround=paving*vec3(0.18,0.195,0.2);
vec3 bankColor=mix(sand,rock,clamp(weights.z*0.7,0.0,0.3));
// Combine coverage before mixing colors so adjoining materials fully hide Mars.
float groundCoverage=clamp(foundation+growth,0.0,1.0);
float natureWeight=growth*(1.0-foundation*0.85);
float natureRatio=natureWeight/max(foundation+natureWeight,0.0001);
base=mix(base,mix(preparedGround,nature,natureRatio),groundCoverage);
base=mix(base,bankColor,bank);
base=mix(base,mud,weights.w*0.55);
float visible=max(groundCoverage,1.0-smoothstep(bankWidth,bankWidth+0.012,shore));
diffuseColor*=vec4(base,smoothstep(0.001,0.02,visible));
// Evaluate derivatives before alpha testing discards neighboring fringe fragments.
vec3 terrainDx=dFdx(-vViewPosition),terrainDy=dFdy(-vViewPosition);
vec2 terrainUvDx=dFdx(vBoardPos),terrainUvDy=dFdy(vBoardPos);
float pavingHeight=aggregate*0.00004*foundation;
vec2 pavingGradient=vec2(dFdx(pavingHeight),dFdy(pavingHeight));
float basinShore=landscapeBasinShore(vFieldLocal,vFieldLayer,vFieldBirth);
#ifdef LANDSCAPE_BASIN
  if(basinShore>=uBasinReach.y) {discard;}
  diffuseColor.a=1.0;
#else
  if(basinShore<uBasinReach.y) {discard;}
#endif
