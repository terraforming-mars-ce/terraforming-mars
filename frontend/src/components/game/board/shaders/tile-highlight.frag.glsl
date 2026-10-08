uniform sampler2D uTileHighlightData;
uniform vec4 uTileHighlightGrid;
uniform vec2 uTileHighlightProjection;
uniform mat4 uTileHighlightBodyInverse;
uniform float uTileHighlightTime;
uniform float uTileHighlightActive;
varying vec3 vTileHighlightWorldPosition;
vec4 tileHighlightTexel(vec2 cell, float offset) {
  vec2 dimensions=vec2(uTileHighlightGrid.z*5.0,uTileHighlightGrid.w);
  return texture2D(uTileHighlightData,(vec2(cell.x*5.0+offset,cell.y)+0.5)/dimensions);
}
vec3 applyTileHighlight(vec3 color) {
  if(uTileHighlightActive<0.5) { return color; }
  vec3 local=(uTileHighlightBodyInverse*vec4(vTileHighlightWorldPosition,1.0)).xyz;
  vec3 direction=normalize(local);
  if(direction.z<=0.0) { return color; }
  float d=length(direction.xy);
  vec2 board=direction.xy*(acos(clamp(direction.z,-1.0,1.0))*uTileHighlightProjection.x/max(d,0.0000001));
  float r=board.y*uTileHighlightProjection.y/0.28274333882;
  float q=board.x/0.32648388556-r*0.5;
  vec3 cube=vec3(q,r,-q-r);
  vec3 rounded=floor(cube+0.5), error=abs(rounded-cube);
  if(error.x>error.y&&error.x>error.z) { rounded.x=-rounded.y-rounded.z; }
  else if(error.y>error.z) { rounded.y=-rounded.x-rounded.z; }
  vec2 cell=rounded.xy-uTileHighlightGrid.xy;
  if(any(lessThan(cell,vec2(0.0)))||any(greaterThanEqual(cell,uTileHighlightGrid.zw))) { return color; }
  vec4 vp=tileHighlightTexel(cell,1.0);
  if(vp.a<0.5) { return color; }
  vec4 state=tileHighlightTexel(cell,0.0);
  if(state.a<=0.002||state.r+state.g+state.b<=0.0) { return color; }
  vec4 normal=tileHighlightTexel(cell,2.0);
  float facing=dot(direction,normal.xyz);
  if(facing<=0.0) { return color; }
  vec3 delta=direction*(normal.w/facing)-normal.xyz*normal.w;
  vec2 p=vec2(dot(delta,tileHighlightTexel(cell,3.0).xyz),dot(delta,tileHighlightTexel(cell,4.0).xyz));
  float edge=max(abs(p.x),max(abs(0.5*p.x+0.86602540378*p.y),abs(-0.5*p.x+0.86602540378*p.y)));
  float mask=1.0-smoothstep(0.14176021702,0.14376021702,edge);
  float gradient=0.25+0.75*smoothstep(0.15,0.45,length(p)/0.332);
  float selected=step(1.5,state.g);
  float pulse=mix(0.5+0.3*sin(uTileHighlightTime*2.0),0.8+0.2*sin(uTileHighlightTime*6.0),selected);
  float placement=min(state.g,1.0)*pulse*mix(0.6,1.0,selected);
  float strength=state.r+placement+state.b;
  vec3 placementTint=mix(vec3(0.4,1.0,0.4),vec3(0.8,1.0,0.8),selected);
  vec3 tint=(vec3(0.95,0.95,1.0)*state.r+placementTint*placement+vp.rgb*state.b)/max(strength,0.00001);
  return mix(color,tint,min(mix(0.55,0.75,selected),strength*gradient)*mask*state.a);
}
