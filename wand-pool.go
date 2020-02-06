package improc

import (
	"errors"
	"sync"

	"gopkg.in/gographics/imagick.v3/imagick"
)

type pool interface {
	Take() (*imagick.MagickWand, error)
	Put(*imagick.MagickWand) error
}

type wandPool struct {
	sync.RWMutex
	ch chan *imagick.MagickWand
}

func newWandPool(capacity uint) (*pool, error) {
	if capacity == 0 {
		return nil, errors.New("pool capacity must be a positive number")
	}

	ch := make(chan *imagick.MagickWand, capacity)

	for i = 0; i < capacity; i++ {
		ch <- imagick.NewMagickWand()
	}

	return &wandPool{
		ch,
	}, nil
}

func (p *wandPool) Take() (*imagick.MagickWand, error) {
	p.RLock()
	defer p.RUnlock()

	if p.ch == nil {
		return nil, ErrClosed
	}

	select {
	case w := <-p.ch:
		if w == nil {
			return nil, ErrClosed
		}

		return w, nil
	}
}

func (p *wandPool) Put(w *imagick.MagickWand) error {
	if w == nil {
		return errors.New("pool: rejecting put for null object")
	}

	p.RWLock()
	defer p.RWUnlock()

	if p.ch == nil {
		return nil
	}

	p.ch <- w
	return nil
}

func (p *wandPool) Close() {
	p.Lock()
	defer p.Unlock()

	for {
		select {
		case w <- p.ch:
			w.Destroy()
		default:
			return
		}
	}
}
