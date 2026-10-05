uniform sampler2D uMars;
uniform sampler2D uDryGrass;
uniform sampler2DArray uGroundAlbedo;
uniform sampler2DArray uGroundDetail;
uniform sampler2D uSand;
uniform sampler2D uRock;
uniform sampler2D uPaving;
uniform sampler2D uPlantShade;
uniform vec4 uPlantShadeBounds;
uniform vec2 uBeachWidths;
uniform float uFrost;
uniform float uIce;
uniform float uGreening;
uniform float uMeadow;
uniform sampler2D uGroundBakeAlbedo;
uniform sampler2D uGroundBakeSurface;
uniform vec4 uGroundBakeBounds;
uniform float uGroundBaked;
varying vec2 vBoardPos;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
// Layers: 0 lush, 1 litter, 2 mud, 3 frost, 4 patchy frost.
// Explicit gradients let the body skip layers per pixel without breaking mip selection.
vec4 groundLayer(sampler2DArray layers,vec2 uv,float layer,vec2 dx,vec2 dy) {
  return textureGrad(layers,vec3(uv.x,-uv.y,layer),vec2(dx.x,-dx.y),vec2(dy.x,-dy.y));
}
vec4 boardTexture(sampler2D map,float scale,vec2 offset,vec2 dx,vec2 dy) {
  return textureGrad(map,vBoardPos*scale+offset,dx*scale,dy*scale);
}
//#pragma body
vec4 field=landscapeField(vFieldLocal,vFieldLayer,vFieldBirth);
float shore=field.y;
float foundation=field.w;
// Everything needing neighbouring fragments happens before the ownership discards below.
vec3 terrainDx=dFdx(-vViewPosition),terrainDy=dFdy(-vViewPosition);
vec2 terrainUvDx=dFdx(vBoardPos),terrainUvDy=dFdy(vBoardPos);
vec2 materialUv=vBoardPos*18.0;
float detailFootprint=length(fwidth(materialUv));
vec3 paving=texture2D(uPaving,vBoardPos*9.0).rgb;
float aggregate=dot(paving,vec3(0.2126,0.7152,0.0722));
float pavingHeight=aggregate*0.00004*foundation;
vec2 pavingGradient=vec2(dFdx(pavingHeight),dFdy(pavingHeight));
vec2 uvDx=terrainUvDx,uvDy=terrainUvDy;
vec3 sphereDirection=normalize(landscapeProject(vBoardPos,0.0));
vec2 marsUv=vec2(atan(sphereDirection.z,-sphereDirection.x)/(2.0*PI),acos(clamp(sphereDirection.y,-1.0,1.0))/PI);
marsUv.y=1.0-marsUv.y;
vec2 marsDx=dFdx(marsUv),marsDy=dFdy(marsUv);
vec2 bakeScale=1.0/uGroundBakeBounds.zw;
vec2 bakeUv=(vBoardPos-uGroundBakeBounds.xy)*bakeScale;
// Outputs of the material stage, either recomposed here or read from the ground bake.
vec3 base;
float groundAlpha;
vec2 surfaceSlope;
float surfaceRoughness;
float landscapeAO;
#ifdef GROUND_BAKE
  bool groundBaked=false;
#else
  // Pixels outside the baked region (or before any bake) run the live path. Both paths sample with
  // explicit gradients, so the per-pixel branch is safe.
  bool groundBaked=uGroundBaked>0.5&&all(greaterThan(bakeUv,vec2(0.0005)))&&all(lessThan(bakeUv,vec2(0.9995)));
#endif
if(groundBaked) {
  // The bake already resolved ownership per texel; only the lake reach needs the exact field.
  #ifdef LANDSCAPE_BASIN
    if(shore>=uBasinReach.y) {discard;}
  #else
    if(shore<uBasinReach.y) {discard;}
  #endif
  vec4 baked=textureGrad(uGroundBakeAlbedo,bakeUv,uvDx*bakeScale,uvDy*bakeScale);
  // Both targets are premultiplied by the ground's alpha (see the bake pass).
  vec4 surface=textureGrad(uGroundBakeSurface,bakeUv,uvDx*bakeScale,uvDy*bakeScale)/max(baked.a,0.0001);
  base=baked.rgb/max(baked.a,0.0001);
  groundAlpha=baked.a;
  surfaceSlope=surface.xy*2.0-1.0;
  surfaceRoughness=surface.z;
  landscapeAO=surface.w;
} else {
  vec4 weights=landscapeMaterials(vFieldLocal,vFieldLayer,vFieldBirth);
  vec4 detailField=landscapeDetail(vFieldLocal,vFieldLayer,vFieldBirth);
  vec3 marsSample=textureGrad(uMars,marsUv,marsDx,marsDy).rgb;
  base=marsClimateSurface(marsSample,marsUv);
  // Preserve contrast throughout the soil's dark features instead of flattening them at a cutoff.
  float marsRelief=pow(clamp(marsSample.r/0.65,0.0,1.0),1.15);
  float naturalShade=mix(0.30,1.0,marsRelief);
  // The ground and basin passes draw the same patches and each owns one side of the lake reach,
  // so the other side is dropped before any of the texturing below.
  float basinShore=landscapeBasinShore(vFieldLocal,vFieldLayer,vFieldBirth);
  #ifdef LANDSCAPE_BASIN
    if(basinShore>=uBasinReach.y) {discard;}
  #else
    if(basinShore<uBasinReach.y) {discard;}
  #endif
  float coverage=clamp(field.z,0.0,1.0);
  float growth=mix(coverage,1.0-pow(1.0-coverage,2.2),smoothstep(0.2,0.7,coverage));
  float bankWidth=mix(uBeachWidths.y,uBeachWidths.x,clamp(weights.y+weights.z,0.0,1.0));
  float bank=1.0-smoothstep(bankWidth*0.55,bankWidth,max(shore,0.0));
  // Combine coverage before mixing colors so adjoining materials fully hide Mars.
  float groundCoverage=clamp(foundation+growth,0.0,1.0);
  float natureWeight=growth*(1.0-foundation*0.85);
  float natureRatio=natureWeight/max(foundation+natureWeight,0.0001);
  mat2 lushTurn=mat2(0.8,-0.6,0.6,0.8);
  vec2 lushUv=lushTurn*materialUv*0.83;
  vec2 lushDx=lushTurn*uvDx*(18.0*0.83),lushDy=lushTurn*uvDy*(18.0*0.83);
  vec2 litterUv=materialUv*0.7;
  vec2 soilUv=materialUv*0.65;
  // Only layers present at this pixel are sampled.
  vec4 present=step(0.001,weights);
  vec4 lushDetail=vec4(0.5,0.5,0.9,0.5);
  vec4 litterDetail=vec4(0.5,0.5,0.9,0.5);
  vec4 soilDetail=vec4(0.5,0.5,0.9,0.5);
  if(present.y>0.0) {lushDetail=groundLayer(uGroundDetail,lushUv,0.0,lushDx,lushDy);}
  if(present.z>0.0) {litterDetail=groundLayer(uGroundDetail,litterUv,1.0,uvDx*12.6,uvDy*12.6);}
  if(present.w>0.0) {soilDetail=groundLayer(uGroundDetail,soilUv,2.0,uvDx*11.7,uvDy*11.7);}
  // Height-lerp splatting: where layers meet, the one whose relief stands higher wins, giving
  // crisp terrain-like transitions instead of a soft cross-fade.
  vec4 layerHeights=vec4(0.5,lushDetail.a,litterDetail.a,soilDetail.a);
  vec4 lifted=(weights+layerHeights*0.5)*present;
  float crest=max(max(lifted.x,lifted.y),max(lifted.z,lifted.w))-0.2;
  vec4 materialWeights=max(lifted-crest,0.0)*present;
  materialWeights/=max(0.0001,dot(materialWeights,vec4(1.0)));
  vec3 mud=vec3(0.0);
  if(present.w>0.0) {mud=groundLayer(uGroundAlbedo,soilUv,2.0,uvDx*11.7,uvDy*11.7).rgb*vec3(0.65,0.7,0.6);}
  vec3 nature=vec3(0.0);
  float macro=0.5;
  float plantShade=0.0;
  if(natureWeight>0.0) {
    vec3 dry=vec3(0.0),lush=vec3(0.0),litter=vec3(0.0);
    if(materialWeights.x>0.001) {dry=boardTexture(uDryGrass,80.0,vec2(0.0),uvDx,uvDy).rgb*vec3(0.48,0.66,0.32);}
    if(present.y>0.0) {lush=groundLayer(uGroundAlbedo,lushUv,0.0,lushDx,lushDy).rgb*vec3(0.65,1.1,0.6);}
    if(present.z>0.0) {litter=groundLayer(uGroundAlbedo,litterUv,1.0,uvDx*12.6,uvDy*12.6).rgb*vec3(0.48,0.58,0.4);}
    nature=dry*materialWeights.x+lush*materialWeights.y+litter*materialWeights.z+mud*materialWeights.w;
    // Greenery ground follows the climate: frozen tundra, brown, brown-green, dark green, then
    // grass. Patches run ahead or behind (wet ground first) so each step spreads organically.
    float greenPatch=boardTexture(uClimateNoise,3.7,vec2(0.11),uvDx,uvDy).r*0.6+lushDetail.a*0.4;
    float groundStage=clamp(uGreening*4.0+(greenPatch-0.55)*0.8+weights.w*0.3,0.0,4.0);
    vec3 groundTone=mix(vec3(0.36,0.37,0.38),vec3(0.3,0.23,0.15),clamp(groundStage,0.0,1.0));
    groundTone=mix(groundTone,vec3(0.22,0.22,0.12),clamp(groundStage-1.0,0.0,1.0));
    groundTone=mix(groundTone,vec3(0.11,0.19,0.08),clamp(groundStage-2.0,0.0,1.0));
    float toneDetail=0.12+0.16*dot(layerHeights.yw,vec2(0.6,0.4));
    nature=mix(groundTone*(0.55+toneDetail*1.8),nature,smoothstep(3.0,4.0,groundStage));
    // Large-scale variation breaks up the tiling: drier, paler patches and darker, damper hollows.
    macro=boardTexture(uClimateNoise,2.4,vec2(0.0),uvDx,uvDy).r;
    nature*=mix(0.84,1.12,macro);
    nature=mix(nature,nature*vec3(1.04,1.0,0.9),smoothstep(0.55,0.8,macro)*0.25);
    // Hill tops catch more light, hollows between them sit in shade.
    float hill=clamp((detailField.g-0.25)/0.45,0.0,1.0);
    nature*=mix(1.0,mix(0.84,1.06,hill),detailField.b);
    nature*=naturalShade;
    vec2 shadeScale=1.0/uPlantShadeBounds.zw;
    plantShade=textureGrad(uPlantShade,(vBoardPos-uPlantShadeBounds.xy)*shadeScale,uvDx*shadeScale,uvDy*shadeScale).r;
  }
  vec3 bankColor=vec3(0.0);
  float coldMeadow=detailField.r*(1.0-uMeadow);
  if(bank>0.001||coldMeadow>0.001) {
    vec3 sand=boardTexture(uSand,48.0,vec2(0.0),uvDx,uvDy).rgb;
    vec3 rock=boardTexture(uRock,29.6,vec2(0.0),uvDx,uvDy).rgb;
    bankColor=mix(sand,rock,clamp(weights.z*0.7,0.0,0.3));
    bankColor*=mix(0.65,1.0,marsRelief);
  }
  // Shoreline meadows need warmth and water; until then the shore is bare sand.
  nature=mix(nature,bankColor,coldMeadow);
  vec3 preparedGround=paving*vec3(0.18,0.195,0.2);
  base=mix(base,mix(preparedGround,nature,natureRatio),groundCoverage);
  base=mix(base,bankColor,bank);
  base=mix(base,mud*naturalShade,weights.w*0.55);
  // Frost settles into hollows first and spares heated city paving.
  float frostAmount=0.0;
  vec4 frostDetail=vec4(0.5,0.5,1.0,0.5);
  if(uFrost>0.0) {
    vec2 frostUv=materialUv*0.45;
    vec2 frostDx=uvDx*(18.0*0.45),frostDy=uvDy*(18.0*0.45);
    frostDetail=groundLayer(uGroundDetail,frostUv,4.0,frostDx,frostDy);
    // A second, rotated sample breaks up the tiling of the hollow mask.
    mat2 frostTurn=mat2(0.6,-0.8,0.8,0.6);
    float hollowB=groundLayer(uGroundDetail,frostTurn*materialUv*0.29+0.37,4.0,frostTurn*uvDx*(18.0*0.29),frostTurn*uvDy*(18.0*0.29)).a;
    float hollow=1.0-(frostDetail.a*0.55+hollowB*0.45);
    // Even at full cold, frost stays a patchy dusting so the ground underneath still reads.
    float threshold=1.0-uFrost*0.8;
    // Frost needs moisture, so it only forms on placed landscape and thins out to nothing at its edge.
    float cover=smoothstep(0.0,0.8,clamp(growth+bank,0.0,1.0)+(hollowB-0.5)*0.3);
    frostAmount=smoothstep(threshold,threshold+0.3,hollow)*(1.0-foundation)*cover;
    // Frozen lakes leave a snow-drift band along the waterline.
    float drift=uIce*(1.0-smoothstep(0.0,0.03,max(shore,0.0)))*smoothstep(0.3,0.6,hollow);
    frostAmount=max(frostAmount,drift*(1.0-foundation));
    if(frostAmount>0.0) {
      float solid=dot(groundLayer(uGroundAlbedo,frostUv,3.0,frostDx,frostDy).rgb,vec3(0.3333));
      vec3 snow=vec3(0.9,0.94,1.0)*(0.92+0.08*solid);
      base=mix(base,snow,frostAmount);
      frostDetail=mix(frostDetail,groundLayer(uGroundDetail,frostUv,3.0,frostDx,frostDy),smoothstep(0.5,0.9,uFrost));
    }
  }
  float visible=max(groundCoverage,1.0-smoothstep(bankWidth,bankWidth+0.012,shore));
  groundAlpha=smoothstep(0.001,0.02,visible);
  float macroRoughness=(macro-0.5)*0.1;
  vec2 lushSlope=mat2(0.8,0.6,-0.6,0.8)*(lushDetail.rg*2.0-1.0);
  surfaceSlope=materialWeights.y*lushSlope+materialWeights.z*(litterDetail.rg*2.0-1.0)+materialWeights.w*(soilDetail.rg*2.0-1.0);
  surfaceSlope=mix(surfaceSlope,frostDetail.rg*2.0-1.0,frostAmount);
  float natureSurface=groundCoverage*natureRatio*(1.0-bank);
  surfaceSlope*=natureSurface;
  float materialRoughness=dot(materialWeights,vec4(0.93,lushDetail.b,litterDetail.b,soilDetail.b));
  materialRoughness=mix(materialRoughness,frostDetail.b,frostAmount);
  surfaceRoughness=mix(0.9,clamp(materialRoughness+macroRoughness,0.65,1.0),natureSurface);
  float cavityHeight=dot(materialWeights,layerHeights);
  cavityHeight=mix(cavityHeight,frostDetail.a,frostAmount);
  float cavityAO=mix(1.0,clamp(cavityHeight*1.3,0.0,1.0),0.45*natureSurface);
  landscapeAO=cavityAO*(1.0-plantShade*0.55*natureSurface)*(1.0-weights.z*0.25);
}
// Both passes classify the same live or baked alpha, after all material derivatives.
#if !defined(GROUND_BAKE) && !defined(LANDSCAPE_BASIN)
  #ifdef LANDSCAPE_EDGE
    if(groundAlpha>=1.0) {discard;}
  #else
    if(groundAlpha<1.0) {discard;}
  #endif
#endif
diffuseColor*=vec4(base,groundAlpha);
#ifdef LANDSCAPE_BASIN
  diffuseColor.a=1.0;
#endif
