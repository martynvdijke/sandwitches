import 'material-dynamic-colors/dist/cdn/material-dynamic-colors.min.js';
import htmx from 'htmx.org';
import 'beercss/dist/cdn/beer.min.css';
import 'beercss/dist/cdn/beer.min.js';
import Chart from 'chart.js/auto';
import Cropper from 'cropperjs';
import 'cropperjs/dist/cropper.min.js';
import 'easymde/dist/easymde.min.css';
import 'easymde/dist/easymde.min.js';
import EasyMDE from 'easymde';
import '../css/zoom.css';
import { initPhotoZoom } from './zoom.js';
// This makes Chart available to your HTML/scripts
window.Chart = Chart;
window.Cropper = Cropper;
window.EasyMDE = EasyMDE;
window.htmx = htmx;
// You can add your own custom JS here
initPhotoZoom();
console.log('Webpack bundle loaded!');
