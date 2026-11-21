// Bake the falloff once. Drawing a handful of these at the emitting particles
// keeps the bright core sharp without blurring the whole canvas every frame.
function glowSprite(color) {
  const sprite = document.createElement("canvas");
  sprite.width = sprite.height = 64;
  const ctx = sprite.getContext("2d");
  const radius = 32;
  const light = ctx.createRadialGradient(
    radius,
    radius,
    0,
    radius,
    radius,
    radius,
  );
  light.addColorStop(0, "#f1fcff");
  light.addColorStop(0.055, `rgba(${color},.98)`);
  light.addColorStop(0.16, `rgba(${color},.52)`);
  light.addColorStop(0.42, `rgba(${color},.15)`);
  light.addColorStop(1, `rgba(${color},0)`);
  ctx.fillStyle = light;
  ctx.fillRect(0, 0, 64, 64);
  return sprite;
}

export function createParticleLight() {
  const sprites = ["138,93,255", "22,189,255", "255,150,67"].map(glowSprite);
  let haze;
  const dust = Array.from({ length: 104 }, (_, i) => ({
    x: ((i + 0.3) * 0.754877666) % 1,
    y: ((i + 0.7) * 0.569840296) % 1,
    depth: ((i + 0.5) * 0.618033989) % 1,
  }));

  function resize(ctx, width, height, scale, dark) {
    if (!scale) return;
    const x = width * 0.5,
      y = height * 0.5;
    haze = ctx.createRadialGradient(x, y, 0, x, y, scale * 1.42);
    haze.addColorStop(0, dark ? "rgba(27,64,155,.26)" : "rgba(41,75,174,.29)");
    haze.addColorStop(0.55, "rgba(44,51,145,.13)");
    haze.addColorStop(1, "rgba(28,54,122,0)");
  }

  function background(ctx, width, height, pan, tilt) {
    ctx.fillStyle = haze;
    ctx.fillRect(0, 0, width, height);
    ctx.fillStyle = "#a6caff";
    for (const star of dust) {
      const edge = Math.min(star.x, 1 - star.x, star.y, 1 - star.y);
      ctx.globalAlpha = (0.15 + 0.38 * star.depth) * Math.min(1, edge * 10);
      const size = 0.55 + 0.9 * star.depth;
      ctx.fillRect(
        star.x * width - pan * 10 * star.depth,
        star.y * height + tilt * 9 * star.depth,
        size,
        size,
      );
    }
    ctx.globalAlpha = 1;
  }

  function glow(ctx, x, y, radius, hue, power) {
    const sprite = sprites[hue >= 9 ? 2 : hue >= 4 ? 1 : 0];
    ctx.globalAlpha = power;
    ctx.drawImage(sprite, x - radius, y - radius, radius * 2, radius * 2);
  }

  return { resize, background, glow };
}
