const { chromium } = require('playwright');
const fs = require('fs'), path = require('path');
const [, , html, outDir, label] = process.argv;
const exe = fs.readdirSync(process.env.HOME + '/Library/Caches/ms-playwright/chromium_headless_shell-1243')
  .map(d => path.join(process.env.HOME, 'Library/Caches/ms-playwright/chromium_headless_shell-1243', d, 'chrome-headless-shell'))
  .find(fs.existsSync);
const VPS = [[320,700],[390,844],[768,1024],[769,1024],[820,1180],[1024,768],[1025,768],[1180,820],[1440,1000]];
(async () => {
  fs.mkdirSync(outDir, { recursive: true });
  const b = await chromium.launch({ executablePath: exe });
  const res = []; let fail = 0;
  for (const [w, h] of VPS) {
    const p = await b.newPage({ viewport: { width: w, height: h } });
    await p.route(/^https?:/, r => (r.request().resourceType() === 'image' && r.request().url().startsWith('https://raw.githubusercontent.com/NeaBouli/TrueRepublic/')) ? r.continue() : r.abort());
    await p.goto('file://' + html); await p.waitForLoadState('networkidle').catch(() => {});
    const r = await p.evaluate(() => {
      const de = document.documentElement, vw = de.clientWidth;
      const links = [...document.querySelectorAll('nav .nav-links a')].map(a => { const b = a.getBoundingClientRect(); const cs = getComputedStyle(a);
        return { t: a.textContent, x: b.left, r: b.right, y: b.top, w: b.width, h: b.height, vis: cs.visibility !== 'hidden' && cs.display !== 'none' && b.width > 0 && b.height > 0 }; });
      const lg = document.querySelector('.logo').getBoundingClientRect();
      links.push({ t: 'LOGO', x: lg.left, r: lg.right, y: lg.top, w: lg.width, h: lg.height, vis: true, logo: true });
      let overlap = 0;
      for (let i = 0; i < links.length; i++) for (let j = i + 1; j < links.length; j++) { const a = links[i], c = links[j];
        if (a.x < c.r && c.x < a.r && a.y < c.y + c.h && c.y < a.y + a.h) overlap++; }
      const heads = [...document.querySelectorAll('h1,h2,h3,h4,h5,h6')].map(e => +e.tagName[1]);
      let skip = 0; for (let i = 1; i < heads.length; i++) if (heads[i] > heads[i - 1] + 1) skip++;
      return { vw, scrollW: de.scrollWidth, links, overlap, mains: document.querySelectorAll('main').length,
        navLabel: document.querySelector('nav')?.getAttribute('aria-label'), h1: heads.filter(x => x === 1).length, skip,
        footerH2: document.querySelectorAll('footer h2').length, footerH4: document.querySelectorAll('footer h4').length,
        iconsHidden: document.querySelectorAll('.feature-icon[aria-hidden="true"]').length,
        useCaseEmojiHidden: document.querySelectorAll('.use-case h3 span[aria-hidden="true"]').length,
        headerPos: getComputedStyle(document.querySelector('header')).position };
    });
    const checks = {
      noHOverflow: r.scrollW <= r.vw,
      navAllVisible: r.links.filter(l => !l.logo).length === 6 && r.links.every(l => l.vis),
      navInViewport: r.links.every(l => l.x >= 0 && l.r <= r.vw + 0.5),
      touch44: r.links.filter(l => !l.logo).every(l => l.h >= 44 && l.w >= 24),
      noLinkOverlap: r.overlap === 0,
      oneMain: r.mains === 1, navLabel: r.navLabel === 'Primary', oneH1: r.h1 === 1, noHeadingSkip: r.skip === 0,
      footerH2: r.footerH2 === 4 && r.footerH4 === 0, iconsDecorative: r.iconsHidden === 11 && r.useCaseEmojiHidden === 3,
    };
    // keyboard: focus order + visible focus ring
    await p.keyboard.press('Tab'); await p.keyboard.press('Tab');
    const f = await p.evaluate(() => { const a = document.activeElement; const cs = getComputedStyle(a); return { t: a.textContent.trim(), ow: cs.outlineWidth, os: cs.outlineStyle }; });
    checks.keyboardFirstNavLink = f.t === 'Features' && f.os !== 'none' && parseFloat(f.ow) >= 2;
    await p.screenshot({ path: `${outDir}/${label}-${w}x${h}-top.png` });
    // anchors: every in-page nav target lands below header, visible
    const anchors = [];
    for (const id of ['features', 'how-it-works', 'rollout', 'license']) {
      await p.evaluate(() => window.scrollTo(0, 0));
      try { await p.click(`nav a[href="#${id}"]`, { timeout: 2000 }); } catch (e) { anchors.push({ id, ok: false, top: NaN, headerBottom: NaN }); continue; }
      await p.waitForTimeout(150);
      const a = await p.evaluate((id) => { const t = document.getElementById(id).getBoundingClientRect(); const hd = document.querySelector('header').getBoundingClientRect();
        const sticky = getComputedStyle(document.querySelector('header')).position === 'sticky';
        return { id, top: t.top, headerBottom: sticky ? hd.bottom : 0, hash: location.hash }; }, id);
      a.ok = a.hash === '#' + id && a.top >= a.headerBottom - 1 && a.top < 200; anchors.push(a);
    }
    checks.anchorsClear = anchors.every(a => a.ok);
    await p.screenshot({ path: `${outDir}/${label}-${w}x${h}-license-anchor.png` });
    await p.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await p.screenshot({ path: `${outDir}/${label}-${w}x${h}-footer.png` });
    const bad = Object.entries(checks).filter(([, v]) => !v).map(([k]) => k);
    if (bad.length) fail++;
    res.push({ vp: `${w}x${h}`, headerPos: r.headerPos, navRows: new Set(r.links.map(l => Math.round(l.y))).size, minLinkH: Math.min(...r.links.filter(l => !l.logo).map(l => l.h)), logoW: Math.round(r.links.find(l => l.logo).w), anchors: anchors.map(a => `${a.id}:${Math.round(a.top)}>=${Math.round(a.headerBottom)}`).join(' '), fail: bad });
    await p.close();
  }
  await b.close();
  console.log(JSON.stringify(res, null, 1)); console.log(label, fail ? `FAIL ${fail}/${VPS.length}` : `PASS ${VPS.length}/${VPS.length}`);
  process.exit(fail ? 1 : 0);
})();
