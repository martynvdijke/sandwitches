// Photo zoom lightbox — vanilla JS, no dependencies.
// Any element with a [data-zoom-src] attribute opens the shared
// #photo-zoom dialog (included once per page via templates/tail.html).
// Event delegation is used so htmx-swapped grid content keeps working.

const MAX_SCALE = 5;
const DBLCLICK_SCALE = 2.5;
const WHEEL_STEP = 1.2;

let dlg = null;
let img = null;
let stage = null;

let scale = 1;
let tx = 0;
let ty = 0;

function apply() {
    if (!img) return;
    img.style.transform = 'translate(' + tx + 'px,' + ty + 'px) scale(' + scale + ')';
}

function clampPan() {
    if (scale <= 1) {
        tx = 0;
        ty = 0;
        return;
    }
    const maxX = (img.clientWidth * (scale - 1)) / 2;
    const maxY = (img.clientHeight * (scale - 1)) / 2;
    tx = Math.min(maxX, Math.max(-maxX, tx));
    ty = Math.min(maxY, Math.max(-maxY, ty));
}

function reset() {
    scale = 1;
    tx = 0;
    ty = 0;
    apply();
}

// Zoom to nextScale keeping the point (cx, cy) visually fixed.
function zoomAt(cx, cy, nextScale) {
    if (!stage) return;
    nextScale = Math.min(MAX_SCALE, Math.max(1, nextScale));
    const rect = stage.getBoundingClientRect();
    const px = cx - rect.left - rect.width / 2;
    const py = cy - rect.top - rect.height / 2;
    const ratio = nextScale / scale;
    tx = px - (px - tx) * ratio;
    ty = py - (py - ty) * ratio;
    scale = nextScale;
    clampPan();
    apply();
}

function open(src, alt) {
    if (!dlg || !src) return;
    img.src = src;
    img.alt = alt || '';
    reset();
    if (typeof dlg.showModal === 'function') {
        dlg.showModal();
    } else {
        dlg.setAttribute('open', '');
    }
}

function close() {
    if (dlg && dlg.open) dlg.close();
}

function initStageInteractions() {
    // Wheel zoom toward cursor.
    stage.addEventListener('wheel', function (e) {
        e.preventDefault();
        const factor = e.deltaY < 0 ? WHEEL_STEP : 1 / WHEEL_STEP;
        zoomAt(e.clientX, e.clientY, scale * factor);
    }, { passive: false });

    // Double-click/double-tap toggles between fit and zoomed.
    stage.addEventListener('dblclick', function (e) {
        e.preventDefault();
        if (scale > 1) {
            reset();
        } else {
            zoomAt(e.clientX, e.clientY, DBLCLICK_SCALE);
        }
    });

    // Pointer-based drag-to-pan and two-finger pinch.
    const pointers = new Map();
    let pinchDist = 0;
    let downX = 0;
    let downY = 0;
    let lastX = 0;
    let lastY = 0;

    function pointerDistance() {
        const pts = Array.from(pointers.values());
        const dx = pts[0].x - pts[1].x;
        const dy = pts[0].y - pts[1].y;
        return Math.hypot(dx, dy);
    }

    stage.addEventListener('pointerdown', function (e) {
        pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
        downX = e.clientX;
        downY = e.clientY;
        if (pointers.size === 2) {
            pinchDist = pointerDistance();
        } else if (pointers.size === 1 && scale > 1) {
            lastX = e.clientX;
            lastY = e.clientY;
            stage.setPointerCapture(e.pointerId);
        }
    });

    stage.addEventListener('pointermove', function (e) {
        if (!pointers.has(e.pointerId)) return;
        pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
        if (pointers.size === 2 && pinchDist > 0) {
            const dist = pointerDistance();
            const pts = Array.from(pointers.values());
            const midX = (pts[0].x + pts[1].x) / 2;
            const midY = (pts[0].y + pts[1].y) / 2;
            zoomAt(midX, midY, scale * (dist / pinchDist));
            pinchDist = dist;
        } else if (pointers.size === 1 && scale > 1) {
            tx += e.clientX - lastX;
            ty += e.clientY - lastY;
            lastX = e.clientX;
            lastY = e.clientY;
            clampPan();
            apply();
        }
    });

    ['pointerup', 'pointercancel'].forEach(function (type) {
        stage.addEventListener(type, function (e) {
            pointers.delete(e.pointerId);
            if (pointers.size < 2) pinchDist = 0;
        });
    });

    // Click on the empty stage around the image acts as backdrop close.
    // Ignore clicks on the image itself and on drag/pinch gestures.
    stage.addEventListener('click', function (e) {
        if (e.target === img) return;
        if (Math.hypot(e.clientX - downX, e.clientY - downY) > 5) return;
        close();
    });
}

export function initPhotoZoom() {
    dlg = document.getElementById('photo-zoom');
    if (!dlg) return;
    img = document.getElementById('photo-zoom-img');
    stage = document.getElementById('photo-zoom-stage');

    document.getElementById('photo-zoom-close').addEventListener('click', close);

    // Click on the backdrop (the dialog itself, outside the image) closes.
    dlg.addEventListener('click', function (e) {
        if (e.target === dlg) close();
    });

    // Reset transform whenever the dialog closes (Escape included).
    dlg.addEventListener('close', reset);

    initStageInteractions();

    // Delegated trigger handling; works for content added later by htmx.
    document.addEventListener('click', function (e) {
        const target = e.target;
        const trigger = target && target.closest ? target.closest('[data-zoom-src]') : null;
        if (!trigger) return;
        e.preventDefault();
        e.stopPropagation();
        open(
            trigger.getAttribute('data-zoom-src'),
            trigger.getAttribute('data-zoom-alt') || trigger.getAttribute('alt') || ''
        );
    });
}
