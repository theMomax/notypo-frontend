package comparison

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIllegalModelInput(t *testing.T) {
	assert.PanicsWithValue(t, ErrIllegalModelInput, func() {
		Compare(stream('a', 'b', 'c', bs), stream('a', 'b', 'c', 'd'), make(chan Comparison, 4))
	})
}

func TestTimeout(t *testing.T) {
	c := make(chan Comparison, 3)
	start := time.Now()
	go Compare(streamuc('a', 'b', 'c'), streamuc('a', 'b', 'c'), c, 500*time.Millisecond)
	consume(c)
	dur := time.Since(start)
	assert.InDelta(t, 500, float64(dur.Nanoseconds())/1.0e6, 20)
}

func TestClosingInputChannel(t *testing.T) {
	c := make(chan Comparison, 0)
	go Compare(stream('a', 'b', 'c', 'd', 'e'), stream('a', 'b', 'c', 'd', 'd', bs, 'e'), c)
	assert.Equal(t, 7, len(consume(c)))
	c = make(chan Comparison, 0)
	go Compare(stream('a', 'b', 'c', 'd'), stream('a', 'b', 'c', 'd', 'd', bs, 'e'), c)
	assert.Equal(t, 4, len(consume(c)))
}

func TestEmptyInputStream(t *testing.T) {
	c := make(chan Comparison, 3)
	go Compare(stream(), stream('a', 'b', 'c'), c)
	assert.Equal(t, 0, len(consume(c)))
	c = make(chan Comparison, 3)
	go Compare(stream('a', 'b', 'c'), stream(), c)
	assert.Equal(t, 0, len(consume(c)))
}

func TestComparisonLogic(t *testing.T) {
	c := make(chan Comparison)
	go Compare(stream(
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', ' ', 'j', 'k', ' ', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	), stream(
		'a', 'b', 'b', '1', ' ', '3', bs, bs, 'c', bs, bs, bs, 'c', 'd', 'e', bs, 'e', 'f', 'g', 'h', 'i', ' ', 'j', 'k', ' ', 'l', 'm', 'n', ' ', bs, 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'x', 'z',
	), c)
	comp := consume(c)
	if len(comp) != 42 {
		assert.FailNow(t, "comparison-stream too short ("+strconv.Itoa(len(comp))+")")
	}
	t.Run("State().Correct()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].State().Correct()
		}, true, true, false, false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true, true, false, false)
	})
	t.Run("State().StatusChanged()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].State().StatusChanged()
		}, false, false, true, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, false, false, false, false, false, false, false, false, false, false, true, false)
	})
	t.Run("len(Changes())", func(t *testing.T) {
		for _, c := range comp {
			assert.Equal(t, 1, len(c.Changes()))
		}
	})
	t.Run("Changes()[0].Rune()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Changes()[0].Rune()
		}, 'a', 'b', 'b', '1', ' ', '3', bs, bs, 'c', bs, bs, bs, 'c', 'd', 'e', bs, 'e', 'f', 'g', 'h', 'i', ' ', 'j', 'k', ' ', 'l', 'm', 'n', ' ', bs, 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'x', 'z')
	})
	t.Run("Changes()[0].Position()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Changes()[0].Position()
		}, 0, 1, 2, 3, 4, 5, 5, 4, 4, 4, 3, 2, 2, 3, 4, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 16, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27)
	})
	t.Run("Changes()[0].Deletion()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Changes()[0].Deletion()
		}, false, false, false, false, false, false, true, true, false, true, true, true, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, false)
	})
	t.Run("Changes()[0].Correct()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Changes()[0].Correct()
		}, true, true, false, false, false, false, true, true, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true, true, true, false, true)
	})
	t.Run("Statistics().TotalCharacters()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Statistics().TotalCharacters()
		}, 1, 2, 3, 4, 5, 6, 5, 4, 5, 4, 3, 2, 3, 4, 5, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28)
	})
	t.Run("Statistics().CorrectCharacters()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Statistics().CorrectCharacters()
		}, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3, 4, 5, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 16, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 26, 26)
	})
	t.Run("Statistics().CorrectWords()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Statistics().CorrectWords()
		}, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2)
	})
	t.Run("Statistics().TotalMisses()", func(t *testing.T) {
		compare(t, func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Statistics().TotalMisses()
		}, 0, 0, 1, 2, 3, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7)
	})
	t.Run("Statistics().FailureRate()", func(t *testing.T) {
		match(t, slice(func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return comp[i].Statistics().FailureRate()
		}), slice(func(i int) interface{} {
			if i < 0 || i >= len(comp) {
				return nil
			}
			return float64(comp[i].Statistics().TotalMisses()) / float64(comp[i].Statistics().TotalCharacters())
		}))
	})
}

func compare(t *testing.T, supplier func(i int) interface{}, expected ...interface{}) {
	match(t, expected, slice(supplier))
}

// match returns true, if the content and length of the given slices match
func match(t *testing.T, slices ...interface{}) {
	if len(slices) < 2 {
		return
	}
	sl := make([][]interface{}, 0, len(slices))
	for _, c := range slices {
		sl = append(sl, c.([]interface{}))
	}
	for _, s := range sl[1:] {
		assert.Equal(t, len(sl[0]), len(s))
	}
	for _, s := range sl[1:] {
		for i, v := range s {
			assert.Equal(t, sl[0][i], v)
		}
	}
}

// stream creates and returns a channel of Characters and channels all the given
// elements into this channel asynchronously
func stream(elements ...rune) chan Character {
	c := make(chan Character, len(elements)/10)
	go func() {
		for _, e := range elements {
			c <- char(e)
		}
		close(c)
	}()
	return c
}

// streamd creates and returns a channel of Characters and channels all the given
// elements into this channel asynchronously and delayed. The delay varies
// randomly between 0 and 50ms
func streamd(elements ...rune) chan Character {
	c := make(chan Character, len(elements)/10)
	go func() {
		for _, e := range elements {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(51)))
			c <- char(e)
		}
		close(c)
	}()
	return c
}

// streamuc creates and returns a channel of Characters and channels all the given
// elements into this channel asynchronously, but won't close the channel
// afterwards
func streamuc(elements ...rune) chan Character {
	c := make(chan Character, len(elements)/10)
	go func() {
		for _, e := range elements {
			c <- char(e)
		}
	}()
	return c
}

// consume channels the given channel into a slice and returns it
func consume(c chan Comparison) (s []Comparison) {
	for {
		v, ok := <-c
		if !ok {
			return
		}
		s = append(s, v)
	}
}

// slice takes a supplier function, that takes an index and returns the element
// which shall be placed at index in the slice returned by this function. The
// slice is completed, when the supplier function returns nil
func slice(supplier func(int) interface{}) (slice []interface{}) {
	slice = make([]interface{}, 0)
	var e interface{}
	i := 0
	for {
		e = supplier(i)
		if e == nil {
			return slice
		}
		slice = append(slice, e)
		i++
	}
}

type char rune

func (c char) Rune() rune {
	return rune(c)
}
