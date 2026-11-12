const fs = require('fs');
const path = require('path');
const { spawn, execSync } = require('child_process');
const puppeteer = require('puppeteer');

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForServer(url, timeoutMs = 15000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(url);
      if (res.ok) return true;
    } catch (e) {
      // wait
    }
    await sleep(250);
  }
  return false;
}

async function main() {
  const imagesDir = path.join(__dirname, '..', 'docs', 'images');
  const framesDir = path.join(imagesDir, 'frames');
  fs.mkdirSync(framesDir, { recursive: true });

  const port = '8089';
  const serverUrl = `http://localhost:${port}`;
  const binPath = path.join(__dirname, '..', 'localecraft.exe');
  const configPath = path.join(__dirname, '..', 'examples', 'localecraft.yaml');

  console.log(`Starting localecraft server on port ${port}...`);
  const serverProc = spawn(binPath, ['serve', '--config', configPath, '--port', port], {
    cwd: path.join(__dirname, '..'),
    stdio: 'ignore',
  });

  const ready = await waitForServer(serverUrl);
  if (!ready) {
    serverProc.kill();
    throw new Error('Server failed to start within timeout');
  }
  console.log('✓ Server is ready at ' + serverUrl);

  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--font-render-hinting=none'],
  });

  try {
    const page = await browser.newPage();

    // 1. Desktop Screenshot (1440x900)
    console.log('Capturing Desktop Screenshot...');
    await page.setViewport({ width: 1440, height: 920, deviceScaleFactor: 2 });
    await page.goto(serverUrl, { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('#root', { timeout: 10000 });
    await sleep(1200);
    await page.screenshot({ path: path.join(imagesDir, 'screenshot_desktop.png') });
    console.log('✓ Wrote screenshot_desktop.png');

    // 2. Tablet Screenshot (768x1024)
    console.log('Capturing Tablet Screenshot...');
    await page.setViewport({ width: 768, height: 1024, deviceScaleFactor: 2 });
    await sleep(600);
    await page.screenshot({ path: path.join(imagesDir, 'screenshot_tablet.png') });
    console.log('✓ Wrote screenshot_tablet.png');

    // 3. Mobile Screenshot (390x844)
    console.log('Capturing Mobile Screenshot...');
    await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
    await sleep(600);
    await page.screenshot({ path: path.join(imagesDir, 'screenshot_mobile.png') });
    console.log('✓ Wrote screenshot_mobile.png');

    // Reset to Desktop for Views & GIF Frames
    await page.setViewport({ width: 1280, height: 820, deviceScaleFactor: 2 });
    await sleep(600);

    // Frame 01: Initial state (nav.save_btn or checkout_btn)
    console.log('Capturing GIF interaction frames...');
    await page.screenshot({ path: path.join(framesDir, 'frame_01.png') });

    // Frame 02: Toggle component to 'destructive' button
    const buttons = await page.$$('button');
    for (const b of buttons) {
      const text = await page.evaluate((el) => el.innerText, b);
      if (text.includes('destructive')) {
        await b.click();
        break;
      }
    }
    await sleep(400);
    await page.screenshot({ path: path.join(framesDir, 'frame_02.png') });

    // Frame 03: Toggle RTL mode
    const rtlBtn = await page.$('button[title*="RTL"]');
    if (rtlBtn) {
      await rtlBtn.click();
      await sleep(400);
    }
    await page.screenshot({ path: path.join(framesDir, 'frame_03.png') });

    // Frame 04: Toggle back and select Matrix tab
    const matrixTab = await page.$('nav button:nth-child(2)');
    if (matrixTab) {
      await matrixTab.click();
      await sleep(500);
    }
    await page.screenshot({ path: path.join(imagesDir, 'screenshot_matrix.png') });
    await page.screenshot({ path: path.join(framesDir, 'frame_04.png') });

    // Frame 05: ICU tab
    const icuTab = await page.$('nav button:nth-child(3)');
    if (icuTab) {
      await icuTab.click();
      await sleep(500);
    }
    await page.screenshot({ path: path.join(imagesDir, 'screenshot_icu.png') });
    await page.screenshot({ path: path.join(framesDir, 'frame_05.png') });

    // Frame 06: Back to Studio
    const studioTab = await page.$('nav button:nth-child(1)');
    if (studioTab) {
      await studioTab.click();
      await sleep(500);
    }
    await page.screenshot({ path: path.join(framesDir, 'frame_06.png') });

    console.log('✓ All frames captured. Compiling GIF with ffmpeg...');
    const gifPath = path.join(imagesDir, 'localecraft_demo.gif');
    const inputPattern = path.join(framesDir, 'frame_%02d.png').replace(/\\/g, '/');
    const paletteFilter = 'fps=1,scale=1200:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse';

    execSync(
      `ffmpeg -y -framerate 0.8 -i "${inputPattern}" -vf "${paletteFilter}" "${gifPath}"`,
      { stdio: 'inherit' }
    );
    console.log('✓ Compiled docs/images/localecraft_demo.gif');

    // Clean up temporary frames folder
    fs.rmSync(framesDir, { recursive: true, force: true });
    console.log('✓ Cleaned up temporary frames');
  } finally {
    await browser.close();
    serverProc.kill();
    console.log('✓ Closed browser and stopped server');
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
