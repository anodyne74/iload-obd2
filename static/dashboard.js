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

    // Mark theme as initialized and update charts if they exist
    themeInitialized = true;
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

    // Update all charts if they exist
    ['rpm-gauge', 'speed-gauge', 'temp-gauge', 
     'rpm-history', 'speed-history', 'temp-history',
     'fuel-map', 'timing-map'].forEach(id => {
        const el = document.getElementById(id);
        if (el && el._fullLayout) {
            Plotly.relayout(id, layout);
        }
    });
}

// Helper function to update history charts
function updateHistoryChart(chartId, data) {
    const el = document.getElementById(chartId);
    if (!el || !el._fullLayout) return;
    
    const trace = {
        y: data,
        x: Array.from({length: data.length}, (_, i) => -i)
    };
    Plotly.update(chartId, trace);
}

// Update engine map visualizations
function updateEngineMap(chartId, data) {
    const el = document.getElementById(chartId);
    if (!el || !el._fullLayout) return;

    // Convert the map data into a format suitable for Plotly heatmap
    const z = data.values;
    const x = data.rpm;
    const y = data.load;

    const trace = {
        x: x,
        y: y,
        z: z,
        type: 'heatmap',
        colorscale: 'Viridis'
    };

    Plotly.update(chartId, trace);
}

// Get description for DTC codes
function getDTCDescription(code) {
    const dtcDescriptions = {
        'P0100': 'Mass or Volume Air Flow Circuit Malfunction',
        'P0101': 'Mass or Volume Air Flow Circuit Range/Performance Problem',
        'P0102': 'Mass or Volume Air Flow Circuit Low Input',
        'P0103': 'Mass or Volume Air Flow Circuit High Input',
        'P0104': 'Mass or Volume Air Flow Circuit Intermittent',
        'P0105': 'Manifold Absolute Pressure/Barometric Pressure Circuit Malfunction',
        'P0106': 'Manifold Absolute Pressure/Barometric Pressure Circuit Range/Performance Problem',
        'P0107': 'Manifold Absolute Pressure/Barometric Pressure Circuit Low Input',
        'P0108': 'Manifold Absolute Pressure/Barometric Pressure Circuit High Input',
        'P0109': 'Manifold Absolute Pressure/Barometric Pressure Circuit Intermittent',
        // Add more DTC codes and descriptions as needed
    };
    return dtcDescriptions[code] || 'Unknown DTC code';
}

// Update dashboard with new telemetry data
function updateDashboard(data) {
    if (!data) return;

    // Update gauges
    if (data.rpm !== undefined) {
        Plotly.update('rpm-gauge', {'value': [data.rpm]});
        rpmHistory.push(data.rpm);
        rpmHistory.shift();
        updateHistoryChart('rpm-history', rpmHistory);
    }

    if (data.speed !== undefined) {
        Plotly.update('speed-gauge', {'value': [data.speed]});
        speedHistory.push(data.speed);
        speedHistory.shift();
        updateHistoryChart('speed-history', speedHistory);
    }

    if (data.temp !== undefined) {
        Plotly.update('temp-gauge', {'value': [data.temp]});
        tempHistory.push(data.temp);
        tempHistory.shift();
        updateHistoryChart('temp-history', tempHistory);
    }

    // Update ECU info
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
            updateEngineMap('fuel-map', data.engineMaps.fuel);
        }
        if (data.engineMaps.timing) {
            updateEngineMap('timing-map', data.engineMaps.timing);
        }
    }

    // Update DTCs
    if (data.dtcs) {
        const dtcList = document.getElementById('dtc-list');
        if (data.dtcs.length === 0) {
            dtcList.innerHTML = 'No DTCs found';
        } else {
            dtcList.innerHTML = data.dtcs.map(dtc => `
                <div class="dtc-item">
                    <span class="dtc-code">${dtc}</span>
                    <span class="dtc-desc">${getDTCDescription(dtc)}</span>
                </div>
            `).join('');
        }
    }
}


// Global variables for charts and data
let ws;
const historyLength = 30;
let rpmHistory = Array(historyLength).fill(0);
let speedHistory = Array(historyLength).fill(0);
let tempHistory = Array(historyLength).fill(0);
let rpmGauge, speedGauge, tempGauge;
let rpmChart, speedChart, tempChart;
let fuelMap, timingMap;

function initDashboard() {
    // Initialize gauges
    rpmGauge = {
        type: 'indicator',
        mode: 'gauge+number',
        value: 0,
        gauge: {
            axis: { range: [0, 8000] },
            bar: { color: '#1e88e5' },
            bgcolor: 'white'
        },
        domain: { row: 0, column: 0 }
    };

    speedGauge = {
        type: 'indicator',
        mode: 'gauge+number',
        value: 0,
        gauge: {
            axis: { range: [0, 200] },
            bar: { color: '#43a047' },
            bgcolor: 'white'
        },
        domain: { row: 0, column: 0 }
    };

    tempGauge = {
        type: 'indicator',
        mode: 'gauge+number',
        value: 0,
        gauge: {
            axis: { range: [0, 150] },
            bar: { color: '#e53935' },
            bgcolor: 'white'
        },
        domain: { row: 0, column: 0 }
    };

    // Initialize history charts
    const historyTrace = {
        y: Array(historyLength).fill(0),
        x: Array.from({length: historyLength}, (_, i) => -i),
        type: 'scatter',
        mode: 'lines',
        line: { width: 2 }
    };

    // Initialize maps
    const mapLayout = {
        xaxis: { title: 'RPM' },
        yaxis: { title: 'Load' }
    };

    // Create all plots
    Plotly.newPlot('rpm-gauge', [rpmGauge], { height: 200, margin: { t: 30, b: 30, l: 30, r: 30 } });
    Plotly.newPlot('speed-gauge', [speedGauge], { height: 200, margin: { t: 30, b: 30, l: 30, r: 30 } });
    Plotly.newPlot('temp-gauge', [tempGauge], { height: 200, margin: { t: 30, b: 30, l: 30, r: 30 } });

    Plotly.newPlot('rpm-history', [{ ...historyTrace, line: { color: '#1e88e5' } }], 
        { height: 150, margin: { t: 10, b: 20, l: 30, r: 10 } });
    Plotly.newPlot('speed-history', [{ ...historyTrace, line: { color: '#43a047' } }],
        { height: 150, margin: { t: 10, b: 20, l: 30, r: 10 } });
    Plotly.newPlot('temp-history', [{ ...historyTrace, line: { color: '#e53935' } }],
        { height: 150, margin: { t: 10, b: 20, l: 30, r: 10 } });

    Plotly.newPlot('fuel-map', [{
        type: 'heatmap',
        colorscale: 'Viridis'
    }], { ...mapLayout, title: 'Fuel Map' });

    Plotly.newPlot('timing-map', [{
        type: 'heatmap',
        colorscale: 'Viridis'
    }], { ...mapLayout, title: 'Timing Map' });

    // Update theme
    updateChartsTheme(document.documentElement.getAttribute('data-theme'));
}

function initWebSocket() {
    // Initialize WebSocket connection
    ws = new WebSocket('ws://' + window.location.host + '/ws');
    
    ws.onmessage = function(evt) {
        const data = JSON.parse(evt.data);
        updateDashboard(data);
    };

    ws.onclose = function() {
        console.log('WebSocket connection closed');
        // Try to reconnect in 5 seconds
        setTimeout(initWebSocket, 5000);
    };

    ws.onerror = function(err) {
        console.error('WebSocket error:', err);
    };
}

// WebSocket initialization
function initWebSocket() {

    // Create engine maps
    fuelMap = create3DMap('fuel-map', 'Fuel Map');
    timingMap = create3DMap('timing-map', 'Timing Map');
}

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
