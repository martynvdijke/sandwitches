// src/static/js/index.js
import 'beercss';
import { ui } from 'beercss';
import 'material-dynamic-colors/dist/cdn/material-dynamic-colors.min.js';
import 'htmx.org/dist/htmx.min.js';

// Ensure beercss ui is global
window.ui = ui;

// Load theme on start
document.addEventListener('DOMContentLoaded', () => {
    const savedMode = localStorage.getItem("mode");
    if (savedMode) {
        ui("mode", savedMode);
    }
});

// Function to toggle mode and save preference
window.toggleMode = () => {
    const currentMode = ui("mode");
    const newMode = currentMode === "dark" ? "light" : "dark";
    ui("mode", newMode);
    localStorage.setItem("mode", newMode);
};

// You can add your own custom JS here
console.log('Webpack bundle loaded!');
