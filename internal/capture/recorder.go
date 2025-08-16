package capture

import (
	"fmt"
	"sync"
)

// Recorder handles the recording of frames to a session
type Recorder struct {
	session  *Session
	running  bool
	mu       sync.Mutex
	handlers map[string]FrameHandler
}

// FrameHandler is an interface for handling different types of frames
type FrameHandler interface {
	HandleFrame(frame Frame) error
	Type() string
}

// NewRecorder creates a new recorder instance
func NewRecorder(vehicleInfo string) *Recorder {
	return &Recorder{
		session:  NewSession(vehicleInfo),
		handlers: make(map[string]FrameHandler),
	}
}

// RegisterHandler adds a frame handler for a specific frame type
func (r *Recorder) RegisterHandler(handler FrameHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[handler.Type()] = handler
}

// Start begins the recording session
func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return fmt.Errorf("recorder is already running")
	}

	r.running = true
	return nil
}

// Stop ends the recording session and saves the data
func (r *Recorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return fmt.Errorf("recorder is not running")
	}

	r.running = false
	return r.session.Save()
}

// Record adds a frame to the current session
func (r *Recorder) Record(frame Frame) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return fmt.Errorf("recorder is not running")
	}

	// Process frame with appropriate handler if available
	if handler, ok := r.handlers[frame.Type]; ok {
		if err := handler.HandleFrame(frame); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}

	r.session.AddFrame(frame)
	return nil
}

// SetMetadata adds metadata to the session
func (r *Recorder) SetMetadata(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.session.SetMetadata(key, value)
}

// IsRunning returns the current recording state
func (r *Recorder) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}
