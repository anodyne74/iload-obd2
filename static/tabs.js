function initTabs() {
    console.log('Initializing tabs...');
    const tabs = document.querySelectorAll('.tab');
    const tabContents = document.querySelectorAll('.tab-content');

    console.log('Found tabs:', tabs.length);
    console.log('Found tab contents:', tabContents.length);

    tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            console.log('Tab clicked:', tab.getAttribute('data-tab'));
            
            // Remove active class from all tabs and contents
            tabs.forEach(t => t.classList.remove('active'));
            tabContents.forEach(c => c.classList.remove('active'));

            // Add active class to clicked tab and corresponding content
            tab.classList.add('active');
            const tabName = tab.getAttribute('data-tab');
            const content = document.getElementById(tabName);
            
            if (content) {
                content.classList.add('active');
                console.log('Activated content:', tabName);
            } else {
                console.error('Could not find content for tab:', tabName);
            }

            // Prevent any default behavior
            e.preventDefault();
            e.stopPropagation();
        });
    });
}

// Enhance the tab styling
const style = document.createElement('style');
style.textContent = `
.tab-content {
    display: none;
    opacity: 0;
    transition: opacity 0.3s ease-in-out;
}

.tab-content.active {
    display: block;
    opacity: 1;
}

.tabs .tab {
    cursor: pointer;
    padding: 10px 20px;
    border-bottom: 2px solid transparent;
    transition: all 0.3s ease-in-out;
}

.tabs .tab:hover {
    background-color: var(--hover-bg-color);
}

.tabs .tab.active {
    border-bottom-color: var(--accent-color);
    color: var(--accent-color);
}

.tabs .tab i {
    margin-right: 8px;
}
`;

document.head.appendChild(style);
