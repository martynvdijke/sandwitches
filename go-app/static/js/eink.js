// ============================================================
// E-Ink Mode JavaScript
// Handles: toggle, cookie persistence, cooking mode
// ============================================================

(function () {
    'use strict';

    const EINK_COOKIE = 'eink_mode';
    const EINK_COOKIE_DAYS = 30;
    const EINK_PARAM = 'eink';

    /**
     * Get cookie value by name
     */
    function getCookie(name) {
        const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
        return match ? decodeURIComponent(match[2]) : null;
    }

    /**
     * Set cookie with expiration
     */
    function setCookie(name, value, days) {
        const d = new Date();
        d.setTime(d.getTime() + days * 24 * 60 * 60 * 1000);
        document.cookie = name + '=' + encodeURIComponent(value) +
            ';expires=' + d.toUTCString() + ';path=/';
    }

    /**
     * Remove cookie
     */
    function eraseCookie(name) {
        document.cookie = name + '=; Max-Age=-99999999; path=/';
    }

    /**
     * Check if e-ink mode is active from URL param
     */
    function isEinkFromUrl() {
        const params = new URLSearchParams(window.location.search);
        return params.get(EINK_PARAM) === '1';
    }

    /**
     * Toggle e-ink mode on/off
     */
    window.toggleEinkMode = function () {
        const isActive = document.body.classList.contains('eink-mode');
        if (isActive) {
            document.body.classList.remove('eink-mode');
            eraseCookie(EINK_COOKIE);
        } else {
            document.body.classList.add('eink-mode');
            setCookie(EINK_COOKIE, '1', EINK_COOKIE_DAYS);
        }
        // Reload to apply server-side template changes
        window.location.reload();
    };

    /**
     * Enable e-ink mode
     */
    window.enableEinkMode = function () {
        document.body.classList.add('eink-mode');
        setCookie(EINK_COOKIE, '1', EINK_COOKIE_DAYS);
    };

    /**
     * Disable e-ink mode
     */
    window.disableEinkMode = function () {
        document.body.classList.remove('eink-mode');
        eraseCookie(EINK_COOKIE);
    };

    // ============================================================
    // Cooking Mode
    // ============================================================

    /**
     * Navigate to a specific cooking step
     */
    window.goToStep = function (step) {
        const params = new URLSearchParams(window.location.search);
        params.set('step', step);
        window.location.search = params.toString();
    };

    /**
     * Go to next cooking step
     */
    window.nextStep = function () {
        const current = parseInt(document.getElementById('cooking-step-input')?.value || '1', 10);
        const total = parseInt(document.getElementById('cooking-step-input')?.dataset?.totalSteps || '1', 10);
        if (current < total) {
            goToStep(current + 1);
        }
    };

    /**
     * Go to previous cooking step
     */
    window.prevStep = function () {
        const current = parseInt(document.getElementById('cooking-step-input')?.value || '1', 10);
        if (current > 1) {
            goToStep(current - 1);
        }
    };

    // On page load, ensure body class matches the eink state
    document.addEventListener('DOMContentLoaded', function () {
        // The server-side sets the body class via template,
        // but this ensures JS toggles stay consistent if cookie changes
        const body = document.body;
        const hasEinkClass = body.classList.contains('eink-mode');
        const cookieActive = getCookie(EINK_COOKIE) === '1';

        if (isEinkFromUrl() || cookieActive) {
            if (!hasEinkClass) {
                body.classList.add('eink-mode');
            }
        }
    });

})();
