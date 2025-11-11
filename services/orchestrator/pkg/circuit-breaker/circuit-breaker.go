package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

type CircuitBreaker struct {
	name        string
	maxFailures uint32
	timeout     time.Duration
	halfOpenMax uint32

	mu            sync.RWMutex
	state         State
	failures      uint32
	successes     uint32
	lastFailTime  time.Time
	halfOpenCount uint32
}

type Config struct {
	Name        string
	MaxFailures uint32
	Timeout     time.Duration
	HalfOpenMax uint32
}

func New(cfg Config) *CircuitBreaker {
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = 5
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.HalfOpenMax == 0 {
		cfg.HalfOpenMax = 1
	}

	return &CircuitBreaker{
		name:        cfg.Name,
		maxFailures: cfg.MaxFailures,
		timeout:     cfg.Timeout,
		halfOpenMax: cfg.HalfOpenMax,
		state:       StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()
	cb.afterRequest(err)
	return err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		if cb.halfOpenCount >= cb.halfOpenMax {
			return ErrTooManyRequests
		}
		cb.halfOpenCount++
		return nil
	default:
		return nil
	}
}

func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.maxFailures {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
		}
	case StateClosed:
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateOpen
		cb.successes = 0
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Name() string {
	return cb.name
}
