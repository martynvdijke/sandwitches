// Browser photo editor — wraps the already-vendored cropperjs v2 web
// components (<cropper-canvas>/<cropper-image>/<cropper-selection> declared
// in templates/admin/recipe_form.html). Zero backend changes: Apply exports
// a JPEG blob into the existing input[name=image] so the normal form POST
// flows through utils.SaveUploadedFile/SaveBytes unchanged.

const MAX_SIDE = 1600; // mirrors utils.MaxImageDimension
const QUALITY = 0.82; // mirrors SaveBytes JPEG quality

let overlay = null;
let fileInput = null;
let previewImg = null;
let canvas = null;
let image = null;
let selection = null;
let currentUrl = '';
let preOpenPreviewSrc = '';
let pendingName = 'edited.jpg';
let lastFocus = null;

// Tracked transform state, re-applied from identity so behavior is
// deterministic regardless of $rotate/$scale relative semantics.
let rot = 0;
let flipX = 1;
let flipY = 1;

function supported() {
    return (
        typeof window.customElements !== 'undefined' &&
        !!window.customElements.get('cropper-canvas') &&
        typeof window.DataTransfer !== 'undefined' &&
        typeof document.createElement('canvas').toBlob === 'function'
    );
}

function fit(w, h, max) {
    const m = Math.max(w, h);
    if (m <= max) return { w: w, h: h };
    return { w: Math.round((w * max) / m), h: Math.round((h * max) / m) };
}

// Decode with EXIF orientation applied and pre-downscale to MAX_SIDE so big
// phone photos neither arrive sideways nor OOM the cropper canvas.
async function normalizedUrl(source) {
    let blob = source instanceof Blob ? source : await (await fetch(source)).blob();
    try {
        if ('createImageBitmap' in window) {
            const bmp = await createImageBitmap(blob, { imageOrientation: 'from-image' });
            try {
                const size = fit(bmp.width, bmp.height, MAX_SIDE);
                const c = document.createElement('canvas');
                c.width = size.w;
                c.height = size.h;
                c.getContext('2d').drawImage(bmp, 0, 0, size.w, size.h);
                const out = await new Promise(function (res) {
                    c.toBlob(res, 'image/jpeg', 0.92);
                });
                if (out) blob = out;
            } finally {
                if (bmp.close) bmp.close();
            }
        }
    } catch (e) {
        // Old browser: fall through with the original blob.
    }
    return URL.createObjectURL(blob);
}

function applyTransform() {
    image.$resetTransform();
    if (rot) image.$rotate(rot + 'deg');
    if (flipX !== 1 || flipY !== 1) image.$scale(flipX, flipY);
}

// $reset() restores the selection captured at element connect time — which
// is zero-area when the overlay is hidden — so size it explicitly instead.
function resetSelection() {
    const w = canvas.clientWidth || 0;
    const h = canvas.clientHeight || 0;
    if (!w || !h) {
        selection.$reset();
        return;
    }
    const bw = Math.round(w * 0.8);
    const bh = Math.round(h * 0.8);
    selection.$change(Math.round((w - bw) / 2), Math.round((h - bh) / 2), bw, bh);
}

function resetState() {
    rot = 0;
    flipX = 1;
    flipY = 1;
    applyTransform();
    resetSelection();
    setAspect(NaN, null);
}

function setAspect(ratio, btn) {
    selection.aspectRatio = ratio;
    overlay.querySelectorAll('[data-aspect]').forEach(function (b) {
        if (btn) b.classList.toggle('fill', b === btn);
        else b.classList.remove('fill');
    });
}

function setImage(url) {
    if (currentUrl) URL.revokeObjectURL(currentUrl);
    currentUrl = url;
    image.setAttribute('src', url);
    return image.$ready().then(function () {
        resetState();
    });
}

async function open(source, name) {
    preOpenPreviewSrc = previewImg ? previewImg.getAttribute('src') : '';
    pendingName = (name || 'edited').replace(/\.[a-z0-9]+$/i, '') + '.jpg';
    lastFocus = document.activeElement;
    overlay.hidden = false;
    document.body.classList.add('editor-open');
    await setImage(await normalizedUrl(source));
    overlay.querySelector('[data-editor-close]').focus();
}

function close() {
    overlay.hidden = true;
    document.body.classList.remove('editor-open');
    if (currentUrl) {
        URL.revokeObjectURL(currentUrl);
        currentUrl = '';
    }
    if (lastFocus && lastFocus.focus) lastFocus.focus();
}

function downscale(source, max) {
    const size = fit(source.width, source.height, max);
    if (size.w === source.width && size.h === source.height) return source;
    const c = document.createElement('canvas');
    c.width = size.w;
    c.height = size.h;
    c.getContext('2d').drawImage(source, 0, 0, size.w, size.h);
    return c;
}

async function apply() {
    if (!selection.width || !selection.height) resetSelection();
    const cropped = await selection.$toCanvas();
    downscale(cropped, MAX_SIDE).toBlob(function (blob) {
        if (!blob) return;
        const dt = new DataTransfer();
        dt.items.add(new File([blob], pendingName, { type: 'image/jpeg' }));
        fileInput.files = dt.files;
        fileInput.dispatchEvent(new Event('change', { bubbles: true }));
        if (previewImg) previewImg.src = URL.createObjectURL(blob);
        const newEdit = overlay.parentElement.querySelector('[data-edit-photo="new"]');
        if (newEdit) newEdit.hidden = false;
        close();
    }, 'image/jpeg', QUALITY);
}

export function initPhotoEditor() {
    fileInput = document.querySelector('input[name="image"][type="file"]');
    overlay = document.getElementById('photo-editor');
    if (!fileInput || !overlay) return;
    if (!supported()) {
        overlay.remove();
        return;
    }
    // Progressive enhancement: Edit buttons ship hidden and are only
    // revealed when the editor can actually run (Cropper + File APIs).
    document.querySelectorAll('[data-edit-photo="existing"]').forEach(function (b) {
        b.hidden = false;
    });
    previewImg = document.getElementById('photo-preview');
    canvas = document.getElementById('pe-canvas');
    image = document.getElementById('pe-image');
    selection = document.getElementById('pe-selection');

    overlay.querySelector('[data-rotate="ccw"]').addEventListener('click', function () {
        rot = (rot - 90) % 360;
        applyTransform();
    });
    overlay.querySelector('[data-rotate="cw"]').addEventListener('click', function () {
        rot = (rot + 90) % 360;
        applyTransform();
    });
    overlay.querySelector('[data-flip="h"]').addEventListener('click', function () {
        flipX *= -1;
        applyTransform();
    });
    overlay.querySelector('[data-flip="v"]').addEventListener('click', function () {
        flipY *= -1;
        applyTransform();
    });
    overlay.querySelector('[data-reset]').addEventListener('click', resetState);
    overlay.querySelectorAll('[data-aspect]').forEach(function (btn) {
        btn.addEventListener('click', function () {
            setAspect(parseFloat(btn.getAttribute('data-aspect')), btn);
        });
    });
    overlay.querySelector('[data-editor-apply]').addEventListener('click', apply);
    overlay.querySelector('[data-editor-close]').addEventListener('click', function () {
        if (previewImg && preOpenPreviewSrc) previewImg.src = preOpenPreviewSrc;
        close();
    });
    overlay.addEventListener('keydown', function (e) {
        if (e.key === 'Escape') {
            e.stopPropagation();
            if (previewImg && preOpenPreviewSrc) previewImg.src = preOpenPreviewSrc;
            close();
        }
    });
    // Backdrop tap closes (ignore taps inside the card / cropper canvas).
    overlay.addEventListener('click', function (e) {
        if (e.target === overlay) {
            if (previewImg && preOpenPreviewSrc) previewImg.src = preOpenPreviewSrc;
            close();
        }
    });

    document.querySelectorAll('[data-edit-photo]').forEach(function (btn) {
        if (btn.getAttribute('data-edit-photo') === 'new') {
            btn.hidden = fileInput.files.length === 0;
            return;
        }
        btn.addEventListener('click', function () {
            const src =
                fileInput.files.length > 0
                    ? fileInput.files[0]
                    : btn.getAttribute('data-full-src') || (previewImg && previewImg.src);
            if (src) open(src, fileInput.files.length > 0 ? fileInput.files[0].name : 'recipe');
        });
    });

    // New-file flow: reveal the Edit button once a file is chosen.
    fileInput.addEventListener('change', function () {
        const btn = document.querySelector('[data-edit-photo="new"]');
        if (!btn) return;
        btn.hidden = fileInput.files.length === 0;
    });
    const newBtn = document.querySelector('[data-edit-photo="new"]');
    if (newBtn) {
        newBtn.addEventListener('click', function () {
            if (fileInput.files.length > 0) open(fileInput.files[0], fileInput.files[0].name);
        });
    }
}
