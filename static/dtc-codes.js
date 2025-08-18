const dtcDescriptions = {
    // Powertrain (P) Codes
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
    'P0110': 'Intake Air Temperature Circuit Malfunction',
    'P0111': 'Intake Air Temperature Circuit Range/Performance Problem',
    'P0112': 'Intake Air Temperature Circuit Low Input',
    'P0113': 'Intake Air Temperature Circuit High Input',
    'P0114': 'Intake Air Temperature Circuit Intermittent',
    'P0115': 'Engine Coolant Temperature Circuit Malfunction',
    'P0116': 'Engine Coolant Temperature Circuit Range/Performance Problem',
    'P0117': 'Engine Coolant Temperature Circuit Low Input',
    'P0118': 'Engine Coolant Temperature Circuit High Input',
    'P0119': 'Engine Coolant Temperature Circuit Intermittent',
    'P0120': 'Throttle/Pedal Position Sensor/Switch A Circuit Malfunction',
    'P0121': 'Throttle/Pedal Position Sensor/Switch A Circuit Range/Performance Problem',
    'P0122': 'Throttle/Pedal Position Sensor/Switch A Circuit Low Input',
    'P0123': 'Throttle/Pedal Position Sensor/Switch A Circuit High Input',
    'P0124': 'Throttle/Pedal Position Sensor/Switch A Circuit Intermittent',
    
    // Hyundai iLoad/H-1 Specific Codes (2.5L CRDi)
    // Fuel and Air System
    'P0087': 'Fuel Rail/System Pressure Too Low - Common Rail System',
    'P0088': 'Fuel Rail/System Pressure Too High - Common Rail System',
    'P0191': 'Fuel Rail Pressure Sensor Circuit Range/Performance',
    'P0192': 'Fuel Rail Pressure Sensor Circuit Low Input',
    'P0193': 'Fuel Rail Pressure Sensor Circuit High Input',
    'P0201': 'Injector Circuit/Open - Cylinder 1',
    'P0202': 'Injector Circuit/Open - Cylinder 2',
    'P0203': 'Injector Circuit/Open - Cylinder 3',
    'P0204': 'Injector Circuit/Open - Cylinder 4',
    'P0234': 'Turbocharger Overboost Condition',
    'P0299': 'Turbocharger Underboost Condition',
    
    // EGR System
    'P0401': 'Exhaust Gas Recirculation Flow Insufficient',
    'P0402': 'Exhaust Gas Recirculation Flow Excessive',
    'P0403': 'Exhaust Gas Recirculation Circuit Malfunction',
    'P0404': 'Exhaust Gas Recirculation Circuit Range/Performance',
    'P0405': 'Exhaust Gas Recirculation Sensor A Circuit Low',
    
    // Diesel Specific
    'P0380': 'Glow Plug/Heater Circuit A Malfunction',
    'P0381': 'Glow Plug/Heater Indicator Circuit Malfunction',
    'P0401': 'Exhaust Gas Recirculation Flow Insufficient',
    'P2002': 'Diesel Particulate Filter Efficiency Below Threshold',
    'P2002': 'Diesel Particulate Filter Missing',
    'P2031': 'Exhaust Gas Temperature Sensor Circuit Low (Bank 1 Sensor 2)',
    'P2033': 'Exhaust Gas Temperature Sensor Circuit High (Bank 1 Sensor 2)',
    
    // Transmission (For automatic transmission models)
    'P0711': 'Transmission Fluid Temperature Sensor Circuit Range/Performance',
    'P0712': 'Transmission Fluid Temperature Sensor Circuit Low Input',
    'P0713': 'Transmission Fluid Temperature Sensor Circuit High Input',
    'P0722': 'Output Speed Sensor Circuit No Signal',
    'P0729': 'Gear 6 Incorrect Ratio',
    'P0731': 'Gear 1 Incorrect Ratio',
    'P0732': 'Gear 2 Incorrect Ratio',
    'P0733': 'Gear 3 Incorrect Ratio',
    'P0734': 'Gear 4 Incorrect Ratio',
    'P0735': 'Gear 5 Incorrect Ratio',
    
    // Vehicle Specific Sensors
    'P0463': 'Fuel Level Sensor Circuit High Input',
    'P0480': 'Cooling Fan 1 Control Circuit Malfunction',
    'P0481': 'Cooling Fan 2 Control Circuit Malfunction',
    'P0500': 'Vehicle Speed Sensor A Malfunction',
    'P0501': 'Vehicle Speed Sensor Range/Performance',
    'P0504': 'Brake Switch A/B Correlation',
    
    // iLoad-Specific Body/Comfort Systems
    'B1060': 'Door Lock Circuit Malfunction',
    'B1350': 'Door Module Failed Programming/Configuration',
    'B1620': 'EEPROM Memory Error',
    'B1701': 'Sliding Door Control Module Fault',
    'B1702': 'Rear Door Control Module Fault',
    
    // Additional Common iLoad Issues
    'P0683': 'Glow Plug Control Module to PCM Communication Circuit',
    'P0684': 'Glow Plug Control Module to PCM Communication Circuit Range/Performance',
    'P0697': 'Sensor Reference Voltage C Circuit High',
    'P2228': 'Barometric Pressure Circuit Low',
    'P2229': 'Barometric Pressure Circuit High',
    
    // Chassis (C) Codes
    'C0035': 'Left Front Wheel Speed Circuit Malfunction',
    'C0040': 'Right Front Wheel Speed Circuit Malfunction',
    'C0045': 'Left Rear Wheel Speed Circuit Malfunction',
    'C0050': 'Right Rear Wheel Speed Circuit Malfunction',
    'C0121': 'Engine Speed Signal Circuit Malfunction',
    
    // Body (B) Codes
    'B1000': 'Climate Control System Malfunction',
    'B1318': 'Battery Voltage Low',
    'B1620': 'EEPROM Error',
    
    // Network (U) Codes
    'U0001': 'High Speed CAN Communication Bus',
    'U0100': 'Lost Communication with ECM/PCM',
    'U0101': 'Lost Communication with TCM',
    'U0121': 'Lost Communication with Anti-Lock Brake System Control Module',
    'U0140': 'Lost Communication with Body Control Module',
    'U0155': 'Lost Communication with Instrument Panel Cluster Control Module'
};
