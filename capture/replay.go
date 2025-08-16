package capture

import (
	"fmt"
	"log"
	"time"
)

type Replayer struct {
	Session      *Session
	Speed        float64 // Replay speed multiplier (1.0 = real-time)
	CurrentFrame int
}

type FrameHandler func(frame CANFrame)

func NewReplayer(session *Session) *Replayer {
	return &Replayer{
		Session:      session,
		Speed:        1.0,
		CurrentFrame: 0,
	}
}

func (r *Replayer) Play(handler FrameHandler) error {
	if len(r.Session.Frames) == 0 {
		return fmt.Errorf("no frames to replay")
	}

	startTime := time.Now()
	sessionStartTime := time.Unix(0, r.Session.Frames[0].Timestamp)

	for i, frame := range r.Session.Frames {
		r.CurrentFrame = i

		// Calculate when this frame should be played
		frameTime := time.Unix(0, frame.Timestamp)
		targetDelay := frameTime.Sub(sessionStartTime)
		actualDelay := time.Since(startTime)

		// Apply speed multiplier
		adjustedDelay := time.Duration(float64(targetDelay) / r.Speed)

		// Wait if we're ahead of schedule
		if actualDelay < adjustedDelay {
			time.Sleep(adjustedDelay - actualDelay)
		}

		handler(frame)
	}

	return nil
}

func (r *Replayer) Pause() {
	// Implement pause functionality
}

func (r *Replayer) Resume() {
	// Implement resume functionality
}

func (r *Replayer) SetSpeed(speed float64) {
	if speed <= 0 {
		log.Printf("Invalid speed multiplier: %v, using 1.0", speed)
		r.Speed = 1.0
		return
	}
	r.Speed = speed
}

func (r *Replayer) JumpTo(timestamp int64) error {
	for i, frame := range r.Session.Frames {
		if frame.Timestamp >= timestamp {
			r.CurrentFrame = i
			return nil
		}
	}
	return fmt.Errorf("timestamp %d not found in session", timestamp)
}

func (r *Replayer) GetProgress() float64 {
	if len(r.Session.Frames) == 0 {
		return 0
	}
	return float64(r.CurrentFrame) / float64(len(r.Session.Frames))
}
