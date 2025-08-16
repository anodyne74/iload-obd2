// Theme handling
function initTheme() {
    const darkModeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const themeSwitch = document.getElementById('theme-switch');
    
    // Set initial theme based on system preference or stored preference
    const savedTheme = localStorage.getItem('theme');
    const isDark = savedTheme === 'dark' || (!savedTheme && darkModeMediaQuery.matches);
    
    document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light');
    themeSwitch.checked = isDark;

    // Handle theme switch changes
    themeSwitch.addEventListener('change', (e) => {
        const theme = e.target.checked ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
        updateChartsTheme(theme);
    });

    // Handle system theme changes
    darkModeMediaQuery.addEventListener('change', (e) => {
        if (!localStorage.getItem('theme')) {
            const theme = e.matches ? 'dark' : 'light';
            document.documentElement.setAttribute('data-theme', theme);
            themeSwitch.checked = e.matches;
            updateChartsTheme(theme);
        }
    });

    // Initial charts theme
    updateChartsTheme(isDark ? 'dark' : 'light');
}

// Update Plotly charts theme
function updateChartsTheme(theme) {
    const isDark = theme === 'dark';
    const chartBg = isDark ? '#242830' : '#ffffff';
    const gridColor = isDark ? '#2c3038' : '#ddd';
    const textColor = isDark ? '#e4e6eb' : '#2c3e50';

    const layout = {
        paper_bgcolor: chartBg,
        plot_bgcolor: chartBg,
        font: { color: textColor },
        xaxis: { gridcolor: gridColor },
        yaxis: { gridcolor: gridColor }
    };

    // Update all charts
    ['rpm-gauge', 'speed-gauge', 'temp-gauge', 
     'rpm-history', 'speed-history', 'temp-history',
     'fuel-map', 'timing-map'].forEach(id => {
        Plotly.relayout(id, layout);
    });
}

// Initialize theme
initTheme();

// Initialize tab functionality
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
        document.querySelectorAll('.tab, .tab-content').forEach(el => el.classList.remove('active'));
        tab.classList.add('active');
        document.getElementById(tab.dataset.tab).classList.add('active');
    });
});

// Initialize WebSocket connection
const ws = new WebSocket('ws://' + window.location.host + '/ws');

// Initialize history data
const historyLength = 30;
const rpmHistory = Array(historyLength).fill(0);
const speedHistory = Array(historyLength).fill(0);
const tempHistory = Array(historyLength).fill(0);

// Create gauges
const rpmGauge = createGauge('rpm-gauge', 'RPM', 0, 6000);
const speedGauge = createGauge('speed-gauge', 'Speed', 0, 180);
const tempGauge = createGauge('temp-gauge', 'Temperature', 0, 150);

// Create history charts
const rpmChart = createHistoryChart('rpm-history', 'RPM History', rpmHistory);
const speedChart = createHistoryChart('speed-history', 'Speed History', speedHistory);
const tempChart = createHistoryChart('temp-history', 'Temperature History', tempHistory);

// Create engine maps
const fuelMap = create3DMap('fuel-map', 'Fuel Map');
const timingMap = create3DMap('timing-map', 'Timing Map');

// Helper functions for visualizations
function createGauge(elementId, title, min, max) {
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const data = [{
        type: 'indicator',
        mode: 'gauge+number',
        value: 0,
        title: { text: title },
        gauge: {
            axis: { 
                range: [min, max],
                tickcolor: isDark ? '#e4e6eb' : '#2c3e50',
            },
            bar: { color: '#e74c3c' },
            bgcolor: isDark ? '#2c3038' : '#f0f0f0',
        }
    }];
    const layout = {
        margin: { t: 25, b: 25, l: 25, r: 25 },
        paper_bgcolor: isDark ? '#242830' : '#ffffff',
        font: { color: isDark ? '#e4e6eb' : '#2c3e50' }
    };
    Plotly.newPlot(elementId, data, layout);
    return { data, layout };
}

function createHistoryChart(elementId, title, data) {
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const trace = {
        y: data,
        type: 'scatter',
        mode: 'lines',
        line: { color: '#e74c3c' }
    };
    const layout = {
        title: title,
        margin: { t: 30, b: 30, l: 30, r: 30 },
        showlegend: false,
        paper_bgcolor: isDark ? '#242830' : '#ffffff',
        plot_bgcolor: isDark ? '#242830' : '#ffffff',
        font: { color: isDark ? '#e4e6eb' : '#2c3e50' },
        xaxis: { 
            gridcolor: isDark ? '#2c3038' : '#ddd',
            zerolinecolor: isDark ? '#2c3038' : '#ddd'
        },
        yaxis: { 
            gridcolor: isDark ? '#2c3038' : '#ddd',
            zerolinecolor: isDark ? '#2c3038' : '#ddd'
        }
    };
    Plotly.newPlot(elementId, [trace], layout);
}

function updateHistory(history, value) {
    history.shift();
    history.push(value);
}

function updateHistoryChart(elementId, data) {
    Plotly.update(elementId, { y: [data] });
}

function create3DMap(elementId, title) {
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    const data = [{
        type: 'surface',
        colorscale: 'Viridis'
    }];
    const layout = {
        title: title,
        autosize: true,
        margin: { l: 30, r: 30, b: 30, t: 30 },
        paper_bgcolor: isDark ? '#242830' : '#ffffff',
        scene: {
            bgcolor: isDark ? '#242830' : '#ffffff',
            xaxis: { 
                gridcolor: isDark ? '#2c3038' : '#ddd',
                zerolinecolor: isDark ? '#2c3038' : '#ddd'
            },
            yaxis: { 
                gridcolor: isDark ? '#2c3038' : '#ddd',
                zerolinecolor: isDark ? '#2c3038' : '#ddd'
            },
            zaxis: { 
                gridcolor: isDark ? '#2c3038' : '#ddd',
                zerolinecolor: isDark ? '#2c3038' : '#ddd'
            }
        },
        font: { color: isDark ? '#e4e6eb' : '#2c3e50' }
    };
    Plotly.newPlot(elementId, data, layout);
    return { data, layout };
}

function updateEngineMap(map, data) {
    const update = {
        z: [data.values],
        x: [data.xAxis],
        y: [data.yAxis]
    };
    Plotly.update(map.elementId, update);
}

// WebSocket message handler
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    // Update live data and gauges
    if (data.rpm) {
        document.getElementById('rpm').textContent = data.rpm;
        rpmGauge.data[0].value = data.rpm;
        updateHistory(rpmHistory, data.rpm);
        Plotly.update('rpm-gauge', rpmGauge.data, rpmGauge.layout);
        updateHistoryChart('rpm-history', rpmHistory);
    }
    if (data.speed) {
        document.getElementById('speed').textContent = data.speed + ' km/h';
        speedGauge.data[0].value = data.speed;
        updateHistory(speedHistory, data.speed);
        Plotly.update('speed-gauge', speedGauge.data, speedGauge.layout);
        updateHistoryChart('speed-history', speedHistory);
    }
    if (data.temp) {
        document.getElementById('temp').textContent = data.temp + ' °C';
        tempGauge.data[0].value = data.temp;
        updateHistory(tempHistory, data.temp);
        Plotly.update('temp-gauge', tempGauge.data, tempGauge.layout);
        updateHistoryChart('temp-history', tempHistory);
    }

    // Update ECU information
    if (data.ecuInfo) {
        document.getElementById('ecu-version').textContent = data.ecuInfo.version || '--';
        document.getElementById('hw-number').textContent = data.ecuInfo.hardware || '--';
        document.getElementById('sw-version').textContent = data.ecuInfo.software || '--';
        document.getElementById('cal-id').textContent = data.ecuInfo.calibration || '--';
        document.getElementById('vin').textContent = data.ecuInfo.vin || '--';
        document.getElementById('build-date').textContent = data.ecuInfo.buildDate || '--';
        document.getElementById('protocol').textContent = data.ecuInfo.protocol || '--';
    }

    // Update engine maps
    if (data.engineMaps) {
        if (data.engineMaps.fuel) {
            updateEngineMap(fuelMap, data.engineMaps.fuel);
        }
        if (data.engineMaps.timing) {
            updateEngineMap(timingMap, data.engineMaps.timing);
        }
    }
    
    // Update DTCs
    if (data.dtcs) {
        const dtcList = document.getElementById('dtc-list');
        if (data.dtcs.length === 0) {
            dtcList.innerHTML = 'No DTCs found';
        } else {
            dtcList.innerHTML = data.dtcs
                .map(dtc => {
                    const description = dtcDescriptions[dtc] || 'Unknown DTC code';
                    return `
                        <div class="dtc-item">
                            <div class="dtc-code">${dtc}</div>
                            <div class="dtc-description">${description}</div>
                        </div>
                    `;
                })
                .join('');
        }
    }
};

// Handle WebSocket disconnection
ws.onclose = function() {
    setTimeout(function() {
        window.location.reload();
    }, 1000);
};
