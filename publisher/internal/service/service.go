package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/SyntropyNet/pubsub-go/pubsub"
	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	Options
	ctx     context.Context
	doneCtx context.Context
}

func New(ctx context.Context, opts ...Option) (*Service, error) {
	ret := &Service{
		ctx: ctx,
	}
	ret.Options.SetDefaults()
	if err := ret.Options.ParseOptions(opts...); err != nil {
		return nil, err
	}
	if ret.natsSub == nil {
		return nil, fmt.Errorf("subsciber nats must not be nil")
	}
	if ret.natsPub == nil {
		return nil, fmt.Errorf("publisher nats must not be nil")
	}
	return ret, nil
}

// Done returns a context that is triggered when the service is completely shut down.
// This is used for graceful shutdown.
func (s *Service) Done() context.Context {
	return s.doneCtx
}

func (s *Service) publish(msg []byte, subj string) error {
	subject := fmt.Sprintf("%s.%s", s.prefix, subj)
	log.Println("Publishing to:", subject)
	return s.natsPub.Publish(s.ctx, subject, msg)
}

func (s *Service) makeBufferedHandler(rungroup *errgroup.Group, name string, handler analytics.Handler) pubsub.HandlerWithSubject {
	ch := make(chan analytics.Message, s.bufferSize)

	rungroup.Go(func() error {
		for {
			var msg analytics.Message
			select {
			case <-s.ctx.Done():
				return s.ctx.Err()
			case msg = <-ch:
			}
			err := handler(msg, s.publish)
			if err != nil {
				log.Printf("Handler for %s failed: %s", name, err.Error())
				return err
			}
		}
	})

	return func(b []byte, subj string) error {
		msg := analytics.Message{
			Timestamp: time.Now(),
			Subject:   subj,
			Data:      b,
		}

		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case ch <- msg:
		default:
			log.Println("Subject handler buffer overflow")
		}

		return nil
	}
}

// Serve instantiates internal processing pipelines essentially starting the service.
func (s *Service) Serve() {
	rungroup, groupCtx := errgroup.WithContext(s.ctx)
	s.doneCtx = groupCtx

	for subject, handler := range s.analytics.Handlers() {
		s.natsSub.AddHandlerWithSubject(
			subject,
			s.makeBufferedHandler(rungroup, subject, handler),
		)
	}

	rungroup.Go(func() error {
		return s.natsPub.Serve(groupCtx)
	})

	rungroup.Go(func() error {
		return s.natsSub.Serve(groupCtx)
	})

	log.Println("Analytics service started.")

	if err := rungroup.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("service is stopped %s", err.Error())
		}
	}

	var completionGroup errgroup.Group
	completionGroup.Go(func() error {
		return nil
	})
}
