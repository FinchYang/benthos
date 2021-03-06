package writer

import (
	"errors"
	"testing"

	"github.com/Jeffail/benthos/v3/lib/message"
	"github.com/Jeffail/benthos/v3/lib/types"
	"github.com/stretchr/testify/assert"
)

func TestBatchedSendHappy(t *testing.T) {
	parts := []string{
		"foo", "bar", "baz", "buz",
	}

	msg := message.New(nil)
	for _, p := range parts {
		msg.Append(message.NewPart([]byte(p)))
	}

	seen := []string{}
	assert.NoError(t, IterateBatchedSend(msg, func(i int, p types.Part) error {
		assert.Equal(t, i, len(seen))
		seen = append(seen, string(p.Get()))
		return nil
	}))

	assert.Equal(t, parts, seen)
}

func TestBatchedSendALittleSad(t *testing.T) {
	parts := []string{
		"foo", "bar", "baz", "buz",
	}

	msg := message.New(nil)
	for _, p := range parts {
		msg.Append(message.NewPart([]byte(p)))
	}

	errFirst, errSecond := errors.New("first"), errors.New("second")

	seen := []string{}
	err := IterateBatchedSend(msg, func(i int, p types.Part) error {
		assert.Equal(t, i, len(seen))
		seen = append(seen, string(p.Get()))
		if i == 1 {
			return errFirst
		}
		if i == 3 {
			return errSecond
		}
		return nil
	})
	assert.Error(t, err)

	expErr := types.NewBatchError(errFirst).AddErrAt(1, errFirst).AddErrAt(3, errSecond)

	assert.Equal(t, parts, seen)
	assert.Equal(t, expErr, err)
}

func TestBatchedSendFatal(t *testing.T) {
	msg := message.New(nil)
	for _, p := range []string{
		"foo", "bar", "baz", "buz",
	} {
		msg.Append(message.NewPart([]byte(p)))
	}

	seen := []string{}
	err := IterateBatchedSend(msg, func(i int, p types.Part) error {
		assert.Equal(t, i, len(seen))
		seen = append(seen, string(p.Get()))
		if i == 1 {
			return types.ErrTypeClosed
		}
		return nil
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "type was closed")
	assert.Equal(t, []string{"foo", "bar"}, seen)

	seen = []string{}
	err = IterateBatchedSend(msg, func(i int, p types.Part) error {
		assert.Equal(t, i, len(seen))
		seen = append(seen, string(p.Get()))
		if i == 1 {
			return types.ErrNotConnected
		}
		return nil
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "not connected to target source or sink")
	assert.Equal(t, []string{"foo", "bar"}, seen)

	seen = []string{}
	err = IterateBatchedSend(msg, func(i int, p types.Part) error {
		assert.Equal(t, i, len(seen))
		seen = append(seen, string(p.Get()))
		if i == 1 {
			return types.ErrTimeout
		}
		return nil
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "action timed out")
	assert.Equal(t, []string{"foo", "bar"}, seen)
}
