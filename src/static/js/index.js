import 'material-dynamic-colors/dist/cdn/material-dynamic-colors.min.js';
import 'htmx.org/dist/htmx.min.js';
import 'beercss/dist/cdn/beer.min.css';
import 'beercss/dist/cdn/beer.min.js';
import Chart from 'chart.js/auto';
import Cropper from 'cropperjs';
import 'cropperjs/dist/cropper.css';

// This makes Chart available to your HTML/scripts
window.Chart = Chart;
window.Cropper = Cropper;

// You can add your own custom JS here
console.log('Webpack bundle loaded!');
