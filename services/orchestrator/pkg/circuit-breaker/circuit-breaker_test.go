package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := Config{
		Name:        "test-breaker",
		MaxFailures: 3,
		Timeout:     30 * time.Second,
		HalfOpenMax: 2,
	}

	cb := New(cfg)

	assert.NotNil(t, cb)
	assert.Equal(t, "test-breaker", cb.Name())
	assert.Equal(t, StateClosed, cb.State())
	assert.Equal(t, uint32(3), cb.maxFailures)
	assert.Equal(t, 30*time.Second, cb.timeout)
	assert.Equal(t, uint32(2), cb.halfOpenMax)
}

func TestNew_DefaultValues(t *testing.T) {
	cfg := Config{
		Name: "test-breaker",
	}

	cb := New(cfg)

	assert.Equal(t, uint32(5), cb.maxFailures)
	assert.Equal(t, 60*time.Second, cb.timeout)
	assert.Equal(t, uint32(1), cb.halfOpenMax)
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     1 * time.Second,
	})

	err := cb.Execute(func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     1 * time.Second,
	})

	expectedErr := errors.New("operation failed")
	err := cb.Execute(func() error {
		return expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, StateClosed, cb.State()) // Should still be closed (1 failure < 3)
}

func TestCircuitBreaker_StateTransition_ClosedToOpen(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     1 * time.Second,
	})

	// Execute 3 failures to open the circuit
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return errors.New("failure")
		})
		assert.Error(t, err)
	}

	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_Execute_CircuitOpen(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 2,
		Timeout:     100 * time.Millisecond,
	})

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	assert.Equal(t, StateOpen, cb.State())

	// Try to execute - should fail immediately
	err := cb.Execute(func() error {
		return nil // This won't even be called
	})

	assert.Error(t, err)
	assert.Equal(t, ErrCircuitOpen, err)
}

func TestCircuitBreaker_StateTransition_OpenToHalfOpen(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	})

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next request should transition to half-open
	err := cb.Execute(func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.State())
}

func TestCircuitBreaker_HalfOpen_Success(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 5,
	})

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Execute successful requests in half-open state
	// First success - should still be half-open
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.State())

	// Second success - should still be half-open (need 3 successes to close)
	err = cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.State())

	// Third success - should close (3 successes >= maxFailures=3)
	err = cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_HalfOpen_Failure(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	})

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Failure in half-open should immediately open again
	err := cb.Execute(func() error {
		return errors.New("still failing")
	})

	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_ClosedState_ResetOnSuccess(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     1 * time.Second,
	})

	// Fail twice
	cb.Execute(func() error {
		return errors.New("failure")
	})
	cb.Execute(func() error {
		return errors.New("failure")
	})

	// Success should reset failures
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)

	// Failures should be reset
	// Fail twice again - should still be closed
	cb.Execute(func() error {
		return errors.New("failure")
	})
	cb.Execute(func() error {
		return errors.New("failure")
	})

	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 10,
		Timeout:     1 * time.Second,
	})

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 10

	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				cb.Execute(func() error {
					return nil
				})
			}
		}()
	}

	wg.Wait()

	// Should still be in closed state after all operations
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ConcurrentFailures(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 5,
		Timeout:     1 * time.Second,
	})

	var wg sync.WaitGroup
	numGoroutines := 5

	// Launch concurrent failures
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Execute(func() error {
				return errors.New("failure")
			})
		}()
	}

	wg.Wait()

	// Circuit should be open after 5 failures
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_State_ThreadSafe(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 100,
		Timeout:     1 * time.Second,
	})

	var wg sync.WaitGroup
	numGoroutines := 20

	// Concurrently read state
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				state := cb.State()
				assert.Contains(t, []State{StateClosed, StateOpen, StateHalfOpen}, state)
			}
		}()
	}

	wg.Wait()
}

func TestCircuitBreaker_Name(t *testing.T) {
	cb := New(Config{
		Name: "my-circuit-breaker",
	})

	assert.Equal(t, "my-circuit-breaker", cb.Name())
}

func TestCircuitBreaker_OpenState_TimeoutNotReached(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 2,
		Timeout:     100 * time.Millisecond,
	})

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	assert.Equal(t, StateOpen, cb.State())

	// Try immediately - should fail
	err := cb.Execute(func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitOpen, err)

	// Wait a bit but not enough
	time.Sleep(50 * time.Millisecond)

	// Should still be open
	err = cb.Execute(func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, ErrCircuitOpen, err)
}

func TestCircuitBreaker_HalfOpen_MultipleSuccessesRequired(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 5,
	})

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Need 3 successes to close (maxFailures)
	for i := 0; i < 2; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StateHalfOpen, cb.State())
	}

	// Third success should close
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_Execute_ErrorPropagation(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 5,
		Timeout:     1 * time.Second,
	})

	expectedErr := errors.New("custom error")
	err := cb.Execute(func() error {
		return expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestCircuitBreaker_StateConstants(t *testing.T) {
	// Verify state constants are distinct
	assert.NotEqual(t, StateClosed, StateOpen)
	assert.NotEqual(t, StateClosed, StateHalfOpen)
	assert.NotEqual(t, StateOpen, StateHalfOpen)
}

func TestCircuitBreaker_ErrorConstants(t *testing.T) {
	assert.Equal(t, "circuit breaker is open", ErrCircuitOpen.Error())
	assert.Equal(t, "too many requests in half-open state", ErrTooManyRequests.Error())
}

func TestCircuitBreaker_OpenToHalfOpen_ResetsCounters(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 3,
	})

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Transition to half-open should reset halfOpenCount
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.State())

	// Should be able to make multiple requests (up to HalfOpenMax)
	for i := 0; i < 2; i++ {
		err = cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)
	}
}

func TestCircuitBreaker_HalfOpenToOpen_ResetsSuccesses(t *testing.T) {
	cb := New(Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 5,
	})

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Get some successes in half-open
	cb.Execute(func() error {
		return nil
	})
	cb.Execute(func() error {
		return nil
	})

	// Failure should reset successes and open circuit
	err := cb.Execute(func() error {
		return errors.New("failure")
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_ZeroMaxFailures(t *testing.T) {
	cfg := Config{
		Name:        "test",
		MaxFailures: 0,
		Timeout:     1 * time.Second,
	}

	cb := New(cfg)

	// Should use default value
	assert.Equal(t, uint32(5), cb.maxFailures)
}

func TestCircuitBreaker_ZeroTimeout(t *testing.T) {
	cfg := Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     0,
	}

	cb := New(cfg)

	// Should use default value
	assert.Equal(t, 60*time.Second, cb.timeout)
}

func TestCircuitBreaker_ZeroHalfOpenMax(t *testing.T) {
	cfg := Config{
		Name:        "test",
		MaxFailures: 3,
		Timeout:     1 * time.Second,
		HalfOpenMax: 0,
	}

	cb := New(cfg)

	// Should use default value
	assert.Equal(t, uint32(1), cb.halfOpenMax)
}
