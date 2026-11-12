// Site notifications: server-sent events plus in-app toasts and browser
// notifications. Dependency-free and gated by the <meta name="notifications">
// tag emitted in templates/head.html.

const MAX_RECONNECT_ATTEMPTS = 3;
const RECONNECT_DELAY_MS = 3000;
const TOAST_LIFETIME_MS = 8000;

let source = null;
let reconnectTimer = null;
let reconnectAttempts = 0;
let stopped = false;
let initialized = false;

function gateEnabled() {
    const meta = document.querySelector('meta[name="notifications"]');
    return !!meta && meta.getAttribute('content') === '1';
}

function ensureToastContainer() {
    let container = document.getElementById('notification-toasts');
    if (!container) {
        container = document.createElement('div');
        container.id = 'notification-toasts';
        container.className = 'flash-messages notification-toasts';
        document.body.appendChild(container);
    }
    return container;
}

function showToast(title, body, url) {
    const container = ensureToastContainer();
    const toast = document.createElement('div');
    toast.className = 'snackbar active primary';

    const icon = document.createElement('i');
    icon.textContent = 'notifications';
    toast.appendChild(icon);

    const text = document.createElement('div');
    text.className = 'notification-text';
    if (title) {
        const heading = document.createElement('strong');
        heading.textContent = title;
        text.appendChild(heading);
    }
    if (body) {
        const paragraph = document.createElement('span');
        paragraph.textContent = body;
        text.appendChild(paragraph);
    }
    toast.appendChild(text);

    if (url) {
        toast.addEventListener('click', function () {
            window.location.href = url;
        });
    } else {
        toast.style.cursor = 'default';
    }

    container.appendChild(toast);
    window.setTimeout(function () {
        if (toast.parentNode) {
            toast.parentNode.removeChild(toast);
        }
    }, TOAST_LIFETIME_MS);
}

function showBrowserNotification(notification) {
    if (!('Notification' in window) || Notification.permission !== 'granted') {
        return;
    }
    const title = notification.title || 'Sandwitches';
    const options = {
        body: notification.body || '',
        tag: notification.type || 'sandwitches',
        data: { url: notification.url || '/' },
    };
    try {
        const browserNotification = new Notification(title, options);
        browserNotification.onclick = function () {
            window.focus();
            if (notification.url) {
                window.location.href = notification.url;
            }
            browserNotification.close();
        };
    } catch (err) {
        // Some browsers (notably mobile) only allow service-worker
        // notifications; the in-app toast still covers this case.
    }
}

function handleNotification(event) {
    if (!event || typeof event.data !== 'string') {
        return;
    }
    let payload;
    try {
        payload = JSON.parse(event.data);
    } catch (err) {
        return;
    }
    if (!payload || typeof payload !== 'object') {
        return;
    }
    const title = payload.title || 'Notification';
    const body = payload.body || '';
    showToast(title, body, payload.url || '');
    showBrowserNotification(payload);
}

function closeSource() {
    if (source) {
        source.close();
        source = null;
    }
}

function connect() {
    if (stopped || source || !('EventSource' in window)) {
        return;
    }
    try {
        source = new EventSource('/api/v1/notifications/stream', { withCredentials: true });
    } catch (err) {
        source = null;
        return;
    }

    source.addEventListener('notification', handleNotification);

    source.addEventListener('ping', function () {
        reconnectAttempts = 0;
    });

    source.onerror = function () {
        // EventSource retries on its own by default; close it so a rejected
        // session (401/403) cannot spin in an endless reconnect loop. The
        // connection is only retried on the next `online` event.
        closeSource();
    };
}

function onOnline() {
    if (stopped || source) {
        return;
    }
    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        return;
    }
    reconnectAttempts += 1;
    if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
    }
    reconnectTimer = window.setTimeout(function () {
        reconnectTimer = null;
        connect();
    }, RECONNECT_DELAY_MS);
}

function teardown() {
    stopped = true;
    if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
    }
    closeSource();
    window.removeEventListener('online', onOnline);
    window.removeEventListener('pagehide', teardown);
}

function initPermissionButton() {
    const button = document.getElementById('enable-browser-notifications');
    if (!button) {
        return;
    }
    const status = document.getElementById('browser-notification-status');

    if (!('Notification' in window)) {
        button.disabled = true;
        if (status) {
            status.textContent = 'This browser does not support notifications.';
        }
        return;
    }

    const update = function () {
        if (!status) {
            return;
        }
        if (Notification.permission === 'granted') {
            button.disabled = true;
            status.textContent = 'Browser notifications are enabled for this device.';
        } else if (Notification.permission === 'denied') {
            button.disabled = true;
            status.textContent = 'Browser notifications are blocked in your browser settings.';
        } else {
            button.disabled = false;
            status.textContent = 'Notifications are not enabled yet.';
        }
    };

    update();
    button.addEventListener('click', function () {
        Notification.requestPermission().then(update).catch(function () {
            update();
        });
    });
}

export function initNotifications() {
    if (initialized) {
        return;
    }
    initialized = true;

    // The permission button lives on the settings page and must work even when
    // the stream gate is off, so wire it up before checking the gate. It never
    // prompts on page load.
    initPermissionButton();

    if (!gateEnabled()) {
        return;
    }

    window.addEventListener('online', onOnline);
    window.addEventListener('pagehide', teardown);
    connect();
}
